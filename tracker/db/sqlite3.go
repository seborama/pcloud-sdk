package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type SQLite3 struct {
	dbPathFilename string
	db             *sql.DB
}

func NewSQLite3(ctx context.Context, dbPath string) (*SQLite3, error) {
	dbPathFilename := filepath.Join(dbPath, "tracker.db")

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

type Version string

const (
	VersionPrevious Version = "P"
	VersionNew              = "N"
)

type FSEntry struct {
	DeviceID       string // for cloud, this could be used to distinguish multiple accounts on the same cloud provider
	EntryID        uint64
	IsFolder       bool
	IsDeleted      bool
	DeletedFileID  uint64
	Path           string
	Name           string
	ParentFolderID uint64
	Created        time.Time
	Modified       time.Time
	Size           uint64
	Hash           string
}

func (s *SQLite3) AddNewFileSystemEntriesV2(ctx context.Context, fsType FSType, entriesCh <-chan FSEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	for entry := range entriesCh {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO "filesystem"
			(type, version, device_id, entry_id, is_folder, is_deleted, deleted_file_id, path, name, parent_folder_id, created, modified, size, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			fsType,
			VersionNew,
			entry.DeviceID,
			fmt.Sprintf("%d", entry.EntryID),
			entry.IsFolder,
			entry.IsDeleted,
			fmt.Sprintf("%d", entry.DeletedFileID), // TODO: needed?
			entry.Path,
			entry.Name,
			fmt.Sprintf("%d", entry.ParentFolderID),
			entry.Created,
			entry.Modified,
			entry.Size,
			entry.Hash,
		)
		if err != nil {
			return doRollback(tx, errors.WithMessagef(err, "deviceID: %s entryID: %d", entry.DeviceID, entry.EntryID))
		}
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}

type config struct {
	entriesChSize int
}

type Options func(*config)

func WithEntriesChSize(n int) Options {
	return func(obj *config) {
		obj.entriesChSize = n
	}
}

func (s *SQLite3) AddNewFileSystemEntries(ctx context.Context, fsType FSType, opts ...Options) (chan<- FSEntry, <-chan error) {
	cfg := config{
		entriesChSize: 100,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	entriesCh := make(chan FSEntry, cfg.entriesChSize)
	errCh := make(chan error, 0)

	go func() {
		defer close(errCh)

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		for entry := range entriesCh {
			_, err := tx.ExecContext(
				ctx,
				`INSERT INTO "filesystem"
			(type, version, device_id, entry_id, is_folder, is_deleted, deleted_file_id, path, name, parent_folder_id, created, modified, size, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				fsType,
				VersionNew,
				entry.DeviceID,
				fmt.Sprintf("%d", entry.EntryID),
				entry.IsFolder,
				entry.IsDeleted,
				fmt.Sprintf("%d", entry.DeletedFileID), // TODO: needed?
				entry.Path,
				entry.Name,
				fmt.Sprintf("%d", entry.ParentFolderID),
				entry.Created,
				entry.Modified,
				entry.Size,
				entry.Hash,
			)
			if err != nil {
				errCh <- doRollback(tx, errors.WithMessagef(err, "deviceID: %s entryID: %d", entry.DeviceID, entry.EntryID))
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			errCh <- doRollback(tx, err)
			return
		}

		errCh <- nil
		return
	}()

	return entriesCh, errCh
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) getFileSystemEntries(ctx context.Context, fsType FSType, version Version) ([]FSEntry, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
			device_id,
			entry_id,
			is_folder,
			is_deleted,
			deleted_file_id,
			path,
			name,
			parent_folder_id,
			created,
			modified,
			size,
			hash
		 FROM "filesystem"
		 WHERE type = ?
		   AND version = ?`,
		fsType,
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
			&entry.DeviceID,
			&entry.EntryID,
			&entry.IsFolder,
			&entry.IsDeleted,
			&entry.DeletedFileID,
			&entry.Path,
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

func (s *SQLite3) GetPreviousFileSystemEntries(ctx context.Context, fsType FSType) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsType, VersionPrevious)
}

func (s *SQLite3) GetLatestFileSystemEntries(ctx context.Context, fsType FSType) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsType, VersionNew)
}

// FSMutation contains a filesystem mutation: type and details.
type FSMutation struct {
	Type MutationType
	Version
	FSEntry
}

// MutationType describes the type of mutation of a filesystem.
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
					  AND is_deleted = false
					  AND type = '`+string(PCloudFileSystem)+`')

		SELECT
			'`+string(MutationTypeDeleted)+`',
			previous.version,
			previous.device_id,
			previous.entry_id,
			previous.is_folder,
			previous.is_deleted,
			previous.deleted_file_id,
			previous.path,
			previous.name,
			previous.parent_folder_id,
			previous.created,
			previous.modified,
			previous.size,
			previous.hash
		 FROM previous LEFT OUTER JOIN new USING (device_id, entry_id)
		 WHERE new.entry_id IS NULL
		
		 UNION
		
		 SELECT
			'`+MutationTypeCreated+`',
			new.version,
			new.device_id,
			new.entry_id,
			new.is_folder,
			new.is_deleted,
			new.deleted_file_id,
			new.path,
			new.name,
			new.parent_folder_id,
			new.created,
			new.modified,
			new.size,
			new.hash
		 FROM new LEFT OUTER JOIN previous USING (device_id, entry_id)
		 WHERE previous.entry_id IS NULL

		 UNION

		 SELECT
			'`+MutationTypeModified+`',
			new.version,
			new.device_id,
			new.entry_id,
			new.is_folder,
			new.is_deleted,
			new.deleted_file_id,
			new.path,
			new.name,
			new.parent_folder_id,
			new.created,
			new.modified,
			new.size,
			new.hash
		 FROM new JOIN previous USING (device_id, entry_id)
		 WHERE new.parent_folder_id = previous.parent_folder_id
		 	AND (
				-- hash is not relevant for folders and that's just fine
				new.hash != previous.hash
			)

		 UNION

		 SELECT
			'`+MutationTypeMoved+`',
			new.version,
			new.device_id,
			new.entry_id,
			new.is_folder,
			new.is_deleted,
			new.deleted_file_id,
			new.path,
			new.name,
			new.parent_folder_id,
			new.created,
			new.modified,
			new.size,
			new.hash
		  FROM new JOIN previous USING (device_id, entry_id)
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
			&fsMutation.DeviceID,
			&fsMutation.EntryID,
			&fsMutation.IsFolder,
			&fsMutation.IsDeleted,
			&fsMutation.DeletedFileID,
			&fsMutation.Path,
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

type FSType string

const (
	LocalFileSystem  FSType = "local"
	PCloudFileSystem        = "pCloud"
)

func (s *SQLite3) MarkNewFileSystemEntriesAsPrevious(ctx context.Context, fsType FSType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM "filesystem"
		 WHERE version = ?
		   AND type = ?`,
		VersionPrevious,
		fsType,
	)
	if err != nil {
		return doRollback(tx, err)
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE "filesystem"
		 SET version = ?
		 WHERE version = ?
		   AND type = ?`,
		VersionPrevious,
		VersionNew,
		fsType,
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
