package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type SQLite3 struct {
	dbPathFilename string
	db             *sql.DB
}

func NewSQLite3(ctx context.Context, dbPath string) (*SQLite3, error) {
	dbPathFilename := dbPath + "/" + "tracker.db"

	db, err := sql.Open("sqlite3", dbPathFilename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = NewMigrator(db).MigrateUp(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &SQLite3{
		dbPathFilename: dbPathFilename,
		db:             db,
	}, nil
}

type TrackingInfo struct {
	DiffID    uint64
	Timestamp time.Time
}

func (s *SQLite3) GetLatestTrackingInfo(ctx context.Context) (*TrackingInfo, error) {
	ti := TrackingInfo{}
	err := s.db.QueryRowContext(
		ctx,
		`SELECT diff_id, timestamp
		 FROM "tracker_diff"`,
	).Scan(
		&ti.DiffID,
		&ti.Timestamp,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &ti, nil
}

type Version string

const (
	VersionPrevious Version = "P"
	VersionNew              = "N"
)

type FSEntry struct {
	EntryID        uint64
	IsFolder       bool
	IsDeleted      bool
	DeletedFileID  uint64
	Name           string
	ParentFolderID uint64
	Created        time.Time
	Modified       time.Time
	Size           uint64
	Hash           uint64
}

func (s *SQLite3) AddNewFileSystemEntry(ctx context.Context, entry FSEntry) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO "filesystem"
			(version, entry_id, is_folder, is_deleted, deleted_file_id, name, parent_folder_id, created, modified, size, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		VersionNew,
		fmt.Sprintf("%d", entry.EntryID),
		entry.IsFolder,
		entry.IsDeleted,
		fmt.Sprintf("%d", entry.DeletedFileID), // TODO: needed?
		entry.Name,
		fmt.Sprintf("%d", entry.ParentFolderID),
		entry.Created,
		entry.Modified,
		entry.Size,
		fmt.Sprintf("%d", entry.Hash),
	)

	return errors.WithStack(err)
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) getFileSystemEntries(ctx context.Context, version Version) ([]FSEntry, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
			entry_id, is_folder, is_deleted, deleted_file_id, name, parent_folder_id, created, modified, size, hash
		 FROM "filesystem"
		 WHERE version = ?`,
		version,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	fsEntries := []FSEntry{}

	for rows.Next() {
		entry := FSEntry{}
		err := rows.Scan(
			&entry.EntryID,
			&entry.IsFolder,
			&entry.IsDeleted,
			&entry.DeletedFileID,
			&entry.Name,
			&entry.ParentFolderID,
			&entry.Created,
			&entry.Modified,
			&entry.Size,
			&entry.Hash,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fsEntries = append(fsEntries, entry)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fsEntries, nil
}

func (s *SQLite3) GetPreviousFileSystemEntries(ctx context.Context) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, VersionPrevious)
}

func (s *SQLite3) GetLatestFileSystemEntries(ctx context.Context) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, VersionNew)
}

type FSMutation struct {
	Type MutationType
	Version
	FSEntry
}

type MutationType string

const (
	MutationTypeDeleted  MutationType = "deleted"
	MutationTypeCreated               = "created"
	MutationTypeModified              = "modified"
	MutationTypeMoved                 = "moved"
)

func (s *SQLite3) GetPCloudMutations(ctx context.Context) ([]FSMutation, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`WITH previous AS (SELECT * FROM filesystem
						   WHERE version = '`+string(VersionPrevious)+`'
						     AND is_deleted = false),
			  new AS (SELECT * FROM filesystem
					  WHERE version = '`+VersionNew+`'
					  AND is_deleted = false)

		SELECT
			'`+string(MutationTypeDeleted)+`', previous.*
		 FROM previous LEFT OUTER JOIN new USING (entry_id)
		 WHERE new.entry_id IS NULL
		
		 UNION
		
		 SELECT
			'`+MutationTypeCreated+`', new.*
		 FROM new LEFT OUTER JOIN previous USING (entry_id)
		 WHERE previous.entry_id IS NULL

		 UNION

		 SELECT
		 	'`+MutationTypeModified+`', new.*
		 FROM new JOIN previous USING (entry_id)
		 WHERE new.parent_folder_id = previous.parent_folder_id
		 	AND (
				-- hash is 0 for folders but that just fine
				new.hash != previous.hash
			)

		 UNION

		 SELECT
		 	'`+MutationTypeMoved+`', new.*
		 FROM new JOIN previous USING (entry_id)
		 -- it should be noted that a file may both move and change
		 WHERE new.parent_folder_id != previous.parent_folder_id
		 `,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	fsMutations := []FSMutation{}

	for rows.Next() {
		fsMutation := FSMutation{}
		err := rows.Scan(
			&fsMutation.Type,
			&fsMutation.Version,
			&fsMutation.EntryID,
			&fsMutation.IsFolder,
			&fsMutation.IsDeleted,
			&fsMutation.DeletedFileID,
			&fsMutation.Name,
			&fsMutation.ParentFolderID,
			&fsMutation.Created,
			&fsMutation.Modified,
			&fsMutation.Size,
			&fsMutation.Hash,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fsMutations = append(fsMutations, fsMutation)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fsMutations, nil
}

func (s *SQLite3) MarkNewFileSystemEntriesAsPrevious(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM "filesystem" WHERE version = '`+string(VersionPrevious)+`'`,
	)
	if err != nil {
		return doRollback(tx, err)
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE "filesystem" SET version = '`+string(VersionPrevious)+`' WHERE version = '`+string(VersionNew)+`'`,
	)
	if err != nil {
		return doRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}

func doRollback(tx *sql.Tx, err error) error {
	errTx := tx.Rollback()
	if errTx != nil {
		return errors.Wrapf(err, "DB error additionally with failed rollback: %s", errTx.Error())
	}

	return errors.WithStack(err)
}
