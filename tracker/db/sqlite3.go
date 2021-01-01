package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	// sqllite3 sql driver.
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// SQLite3 is a sqlite3 database store.
type SQLite3 struct {
	dbPathFilename string
	db             *sql.DB
}

// NewSQLite3 creates a new initialised SQLite3.
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

// Version is used to distinguish the two entry-sets of file system data in the database.
type Version string

const (
	// VersionPrevious is the previous version of file system entries in the database.
	VersionPrevious Version = "Previous"
	// VersionNew is the newer version of file system entries in the database.
	VersionNew Version = "New"
)

// FSEntry is a set of details about an entry (folder or file) in the file system.
type FSEntry struct {
	FSType         FSType
	DeviceID       string // for cloud, this could be used to distinguish multiple accounts on the same cloud provider
	EntryID        uint64
	IsFolder       bool
	Path           string
	Name           string
	ParentFolderID uint64
	Created        time.Time
	Modified       time.Time
	Size           uint64
	Hash           string
}

type config struct {
	entriesChSize int
}

// Options defines the signature of a functional parameter for AddNewFileSystemEntries.
type Options func(*config)

// WithEntriesChannelSize is a functional parameter that allows to choose the size of the entries
// channel used by AddNewFileSystemEntries.
func WithEntriesChannelSize(n int) Options {
	return func(obj *config) {
		obj.entriesChSize = n
	}
}

// AddNewFileSystemEntries adds a new file system entry.
// It returns two channels:
// - the first is used to supply data to this method
// - the second is a channel of error type should this function encounter an error.
// Refer to tests and main code for example uses.
func (s *SQLite3) AddNewFileSystemEntries(ctx context.Context, opts ...Options) (chan<- FSEntry, <-chan error) {
	cfg := config{
		entriesChSize: 100,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	entriesCh := make(chan FSEntry, cfg.entriesChSize)
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		for entry := range entriesCh {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO "filesystem"
			(type, version, device_id, entry_id, is_folder, path, name, parent_folder_id, created, modified, size, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				entry.FSType,
				VersionNew,
				entry.DeviceID,
				fmt.Sprintf("%d", entry.EntryID),
				entry.IsFolder,
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
	}()

	return entriesCh, errCh
}

// Close close the connection to sqlite3.
func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) getFileSystemEntries(ctx context.Context, fsType FSType, version Version) ([]FSEntry, error) {
	// nolint: rowserrcheck
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
			"type",
			device_id,
			entry_id,
			is_folder,
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
	defer func() { _ = rows.Close() }()

	fsEntries := []FSEntry{}

	for rows.Next() {
		entry := FSEntry{}
		err = rows.Scan(
			&entry.FSType,
			&entry.DeviceID,
			&entry.EntryID,
			&entry.IsFolder,
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

// GetPreviousFileSystemEntries get the previous (i.e. version "Previous") file system entries
// for the specified file system type.
func (s *SQLite3) GetPreviousFileSystemEntries(ctx context.Context, fsType FSType) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsType, VersionPrevious)
}

// GetLatestFileSystemEntries get the latest (i.e. version "New") file system entries for the
// specified file system type.
func (s *SQLite3) GetLatestFileSystemEntries(ctx context.Context, fsType FSType) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsType, VersionNew)
}

// FSMutation contains a file system mutation: type and details.
type FSMutation struct {
	Type MutationType
	Version
	FSEntry
}

type FSMutations []FSMutation

func (fsm FSMutations) Filter(mType MutationType) FSMutations {
	fsMutations := FSMutations{}

	for _, el := range fsm {
		if el.Type == mType {
			fsMutations = append(fsMutations, el)
		}
	}

	return fsMutations
}

// MutationType describes the type of mutation of a file system.
type MutationType string

const (
	// MutationTypeDeleted means a deletion on the file system.
	MutationTypeDeleted MutationType = "deleted"
	// MutationTypeCreated means a creation on the file system.
	MutationTypeCreated MutationType = "created"
	// MutationTypeModified means a file content modification on the file system.
	MutationTypeModified MutationType = "modified"
	// MutationTypeMoved means a file move on the file system.
	MutationTypeMoved MutationType = "moved"
)

// GetPCloudVsLocalMutations returns a slice of mutations that exist between the pCloud file system
// and the local file system.
func (s *SQLite3) GetPCloudVsLocalMutations(ctx context.Context) (FSMutations, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT scm.mutation_type,
				scm.fs_type,
				fs.version,
			    scm.device_id,
			    scm.entry_id,
			    fs.is_folder,
			    fs.path,
			    fs.name,
			    fs.parent_folder_id,
			    fs.created,
			    fs.modified,
			    fs.size,
			    fs.hash
		 FROM staging_cross_mutations scm
			  LEFT OUTER JOIN filesystem fs USING (fs_type, device_id, entry_id)
		 WHERE fs.version = :version_new`,
		sql.Named("version_new", VersionNew),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	fsMutations := FSMutations{}

	for rows.Next() {
		fsMutation := FSMutation{}
		err = rows.Scan(
			&fsMutation.Type,
			&fsMutation.FSType,
			&fsMutation.Version,
			&fsMutation.DeviceID,
			&fsMutation.EntryID,
			&fsMutation.IsFolder,
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

// GetPCloudMutations returns a slice of mutations on the pCloud file system.
// nolint: funlen
func (s *SQLite3) GetPCloudMutations(ctx context.Context) (FSMutations, error) {
	return s.getFileSystemMutations(ctx, PCloudFileSystem)
}

// GetLocalMutations returns a slice of mutations on the local file system.
// nolint: funlen
func (s *SQLite3) GetLocalMutations(ctx context.Context) (FSMutations, error) {
	return s.getFileSystemMutations(ctx, LocalFileSystem)
}

func (s *SQLite3) getFileSystemMutations(ctx context.Context, fsType FSType) (FSMutations, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT scm.mutation_type,
			    scm.fs_type,
			    scm.version,
			    scm.device_id,
			    scm.entry_id,
			    fs.is_folder,
			    fs.path,
			    fs.name,
			    fs.parent_folder_id,
			    fs.created,
			    fs.modified,
			    fs.size,
			    fs.hash
		 FROM staging_fs_mutations scm
			  LEFT OUTER JOIN filesystem fs
			  ON scm.fs_type = fs.type
			  	 AND scm.version = fs.version
			  	 AND scm.device_id = fs.device_id
			  	 AND scm.entry_id = fs.entry_id
		 WHERE scm.fs_type = :fstype`,
		sql.Named("fstype", fsType),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	fsMutations := FSMutations{}

	for rows.Next() {
		fsMutation := FSMutation{}
		err = rows.Scan(
			&fsMutation.Type,
			&fsMutation.FSType,
			&fsMutation.Version,
			&fsMutation.DeviceID,
			&fsMutation.EntryID,
			&fsMutation.IsFolder,
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

// TODO: for this to work reliably, path needs to be standardised (i.e. \ or /, C: or /, etc).
func (s *SQLite3) refreshCrossFSMutationsStagingTable(ctx context.Context, left, right FSType) error {
	_, err := s.db.ExecContext(
		ctx,
		`WITH left_fs AS (SELECT type, device_id, entry_id, path, name, parent_folder_id, hash
			                FROM filesystem
						   WHERE version = :version_new
					         AND type = :left_fs),
			 right_fs AS (SELECT type, device_id, entry_id, path, name, parent_folder_id, hash
				            FROM filesystem
						   WHERE version = :version_new
						     AND type = :right_fs)

		INSERT INTO staging_cross_mutations

		SELECT
			:mutation_type_deleted,
			left_fs.type,
			left_fs.device_id,
			left_fs.entry_id
		 FROM left_fs LEFT OUTER JOIN right_fs USING (path, name)
		 WHERE right_fs.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_created,
			right_fs.type,
			right_fs.device_id,
			right_fs.entry_id
		 FROM right_fs LEFT OUTER JOIN left_fs USING (path, name)
		 WHERE left_fs.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_modified,
			right_fs.type,
			right_fs.device_id,
			right_fs.entry_id
		 FROM left_fs JOIN right_fs USING (path, name)
		 WHERE left_fs.parent_folder_id = right_fs.parent_folder_id
		   AND (
		       -- hash is not relevant for folders and that's just fine
		       left_fs.hash != right_fs.hash
		   )

		 UNION

		 SELECT
			:mutation_type_moved,
			right_fs.type,
			right_fs.device_id,
			right_fs.entry_id
		  FROM left_fs JOIN right_fs USING (path, name)
		 -- it should be noted that a file may both move and change
		 WHERE left_fs.parent_folder_id != right_fs.parent_folder_id`,
		sql.Named("left_fs", left),
		sql.Named("right_fs", right),
		sql.Named("version_new", VersionNew),
		sql.Named("mutation_type_deleted", MutationTypeDeleted),
		sql.Named("mutation_type_created", MutationTypeCreated),
		sql.Named("mutation_type_modified", MutationTypeModified),
		sql.Named("mutation_type_moved", MutationTypeMoved),
	)

	return errors.WithStack(err)
}

func (s *SQLite3) refreshFSMutationsStagingTable(ctx context.Context, fsType FSType) error {
	_, err := s.db.ExecContext(
		ctx,
		`WITH previous AS (SELECT type, version, device_id, entry_id, parent_folder_id, hash
						     FROM filesystem
						    WHERE version = :version_previous
					          AND type = :fstype),
				  new AS (SELECT type, version, device_id, entry_id, parent_folder_id, hash
						    FROM filesystem
						   WHERE version = :version_new
						     AND type = :fstype)

		INSERT INTO staging_fs_mutations

		SELECT
			:mutation_type_deleted,
			previous.type,
			previous.version,
			previous.device_id,
			previous.entry_id
		 FROM previous LEFT OUTER JOIN new USING (device_id, entry_id)
		 WHERE new.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_created,
			new.type,
			new.version,
			new.device_id,
			new.entry_id
		 FROM new LEFT OUTER JOIN previous USING (device_id, entry_id)
		 WHERE previous.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_modified,
			new.type,
			new.version,
			new.device_id,
			new.entry_id
		 FROM new JOIN previous USING (device_id, entry_id)
		 WHERE new.parent_folder_id = previous.parent_folder_id
			AND (
				-- hash is not relevant for folders and that's just fine
				new.hash != previous.hash
			)

		 UNION

		 SELECT
			:mutation_type_moved,
			new.type,
			new.version,
			new.device_id,
			new.entry_id
		  FROM new JOIN previous USING (device_id, entry_id)
		 -- it should be noted that a file may both move and change
		 WHERE new.parent_folder_id != previous.parent_folder_id`,
		sql.Named("fstype", fsType),
		sql.Named("version_previous", VersionPrevious),
		sql.Named("version_new", VersionNew),
		sql.Named("mutation_type_deleted", MutationTypeDeleted),
		sql.Named("mutation_type_created", MutationTypeCreated),
		sql.Named("mutation_type_modified", MutationTypeModified),
		sql.Named("mutation_type_moved", MutationTypeMoved),
	)

	return errors.WithStack(err)
}

// FSType describes a file system type.
type FSType string

const (
	// LocalFileSystem represents the local file system (non-cloud).
	LocalFileSystem FSType = "local"
	// PCloudFileSystem represents the pCloud file system.
	PCloudFileSystem FSType = "pCloud"
)

// SyncStatus defines the status of the sync. It is used to prevent refreshing data in the
// filesystem table when it has not yet been completely sync'ed.
// In particular, VersionPrevious should not be replaced with new data until the sync has
// completed or some delta changes between "previous" and "new" will be lost and non-replicated.
// This should not be confused with the sync that takes place across filesystems (such as cloud
// vs local) which only involves VersionNew of each filesystems.
// It is always safe to update VersionNew (but not VersionPrevious until processed).
type SyncStatus string

const (
	// SyncStatusComplete indicates VersionPrevious has been completely sync'ed and can now be
	// replaced with newer data.
	SyncStatusComplete SyncStatus = "Complete"

	// SyncStatusRequired indicates VersionPrevious has just been refreshed and requires
	// sync'ing against VersionNew.
	SyncStatusRequired SyncStatus = "Required"

	// SyncStatusInProgress indicates that the sync between VersionPrevious and VersionNew is in
	// progress.
	SyncStatusInProgress SyncStatus = "In progress"
)

// DeleteVersionNew removes "new" file system entries for the specified file system.
// This would be performed with a view to load a new "VersionNew" set, in replacement.
func (s *SQLite3) DeleteVersionNew(ctx context.Context, fsType FSType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.deleteVersion(ctx, tx, fsType, VersionNew)
	if err != nil {
		return doRollback(tx, err)
	}

	err = s.deleteFSMutations(ctx, tx, fsType)
	if err != nil {
		return errors.WithStack(err)
	}

	// cross FS mutations are built upon VersionNew so the staging table data needs clearing
	err = s.deleteCrossFSMutations(ctx, tx)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}

func (s *SQLite3) deleteVersionPrevious(ctx context.Context, tx *sql.Tx, fsType FSType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.deleteVersion(ctx, tx, fsType, VersionPrevious)
	if err != nil {
		return doRollback(tx, err)
	}

	err = s.deleteFSMutations(ctx, tx, fsType)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}

func (s *SQLite3) deleteVersion(ctx context.Context, tx *sql.Tx, fsType FSType, version Version) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM "filesystem"
		 WHERE version = ?
		   AND type = ?`,
		version,
		fsType,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLite3) deleteFSMutations(ctx context.Context, tx *sql.Tx, fsType FSType) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM "staging_fs_mutations"
		 WHERE fs_type = ?`,
		fsType,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLite3) deleteCrossFSMutations(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM "staging_cross_mutations"`,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// MarkNewFileSystemEntriesAsPrevious clears the "previous" file system entries for the specified
// file system and marks the "new" file system entries as "previous".
func (s *SQLite3) MarkNewFileSystemEntriesAsPrevious(ctx context.Context, fsType FSType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.deleteVersionPrevious(ctx, tx, fsType)
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

	err = s.deleteFSMutations(ctx, tx, fsType)
	if err != nil {
		return errors.WithStack(err)
	}

	// cross FS mutations are built upon VersionNew so the staging table data needs clearing
	err = s.deleteCrossFSMutations(ctx, tx)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}

// MarkSyncRequired marks the status of the sync as "required".
// This also triggers the internal refresh of all staging tables.
func (s *SQLite3) MarkSyncRequired(ctx context.Context, fsType FSType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, fsType := range []FSType{PCloudFileSystem, LocalFileSystem} {
		err = s.refreshFSMutationsStagingTable(ctx, fsType)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = s.refreshCrossFSMutationsStagingTable(ctx, PCloudFileSystem, LocalFileSystem)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO "sync" ("type", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("type")
		 DO UPDATE SET "status" = excluded.status`,
		fsType,
		SyncStatusRequired,
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

// MarkSyncInProgress marks the status of the sync as "in progress".
func (s *SQLite3) MarkSyncInProgress(ctx context.Context, fsType FSType) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO "sync" ("type", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("type")
		 DO UPDATE SET "status" = excluded.status`,
		fsType,
		SyncStatusInProgress,
	)

	return errors.WithStack(err)
}

// MarkSyncComplete marks the status of the sync as "complete".
// TODO: it may be that the staging table should be cleared down, although not essential because
// that is properly taken care of by other method that change the state of table "filesystem".
func (s *SQLite3) MarkSyncComplete(ctx context.Context, fsType FSType) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO "sync" ("type", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("type")
		 DO UPDATE SET "status" = excluded.status`,
		fsType,
		SyncStatusComplete,
	)

	return errors.WithStack(err)
}

// GetSyncStatus returns the current status of the sync for the specified fsType.
// It will return an error (no rows found) if it cannot find the status row.
func (s *SQLite3) GetSyncStatus(ctx context.Context, fsType FSType) (SyncStatus, error) {
	status := ""

	err := s.db.QueryRowContext(
		ctx,
		`SELECT status FROM "sync" WHERE "type" = ?`,
		fsType,
	).Scan(&status)

	return SyncStatus(status), errors.WithStack(err)
}

// IsFileSystemEmpty returns true if no entry data exists at all in the database for `fsType`,
// otherwise it returns false.
func (s *SQLite3) IsFileSystemEmpty(ctx context.Context, fsType FSType) (bool, error) {
	previousRowsCount := 0

	err := s.db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM "filesystem"
		 WHERE "type" = ?
		   AND "version" = ?`,
		fsType,
		VersionPrevious,
	).Scan(&previousRowsCount)
	if err != nil {
		return false, errors.WithStack(err)
	}

	newRowsCount := 0

	err = s.db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM "filesystem"
		 WHERE "type" = ?
		   AND "version" = ?`,
		fsType,
		VersionNew,
	).Scan(&newRowsCount)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return previousRowsCount == 0 && newRowsCount == 0, nil
}

func doRollback(tx *sql.Tx, err error) error {
	errTx := tx.Rollback()
	if errTx != nil {
		return errors.Wrapf(err, "DB error additionally with failed rollback: %s", errTx.Error())
	}

	return errors.WithStack(err)
}
