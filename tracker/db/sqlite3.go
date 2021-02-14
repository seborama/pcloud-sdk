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

type FSDriver string

const (
	FSDriverPCloud FSDriver = "pCloud"
	FSDriverLocal  FSDriver = "Local"
)

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
	FSName         FSName
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

// GetSyncDetails returns the driver for the specified file system name and the root path.
func (s *SQLite3) GetSyncDetails(ctx context.Context, fsName FSName) (FSDriver, string, error) {
	var fsDriver FSDriver
	fsRoot := ""

	err := s.db.QueryRowContext(
		ctx,
		`SELECT fs_driver, fs_root FROM "sync" WHERE "fs_name" = ?`,
		fsName,
		fsRoot,
	).Scan(&fsDriver, &fsRoot)

	return fsDriver, fsRoot, errors.WithStack(err)
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
			(fs_name, version, device_id, entry_id, is_folder, path, name, parent_folder_id, created, modified, size, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				entry.FSName,
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

// Close closes the connection to sqlite3.
func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) getFileSystemEntries(ctx context.Context, fsName FSName, version Version) ([]FSEntry, error) {
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
		fsName,
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
			&entry.FSName,
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
func (s *SQLite3) GetPreviousFileSystemEntries(ctx context.Context, fsName FSName) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsName, VersionPrevious)
}

// GetLatestFileSystemEntries get the latest (i.e. version "New") file system entries for the
// specified file system type.
func (s *SQLite3) GetLatestFileSystemEntries(ctx context.Context, fsName FSName) ([]FSEntry, error) {
	return s.getFileSystemEntries(ctx, fsName, VersionNew)
}

// FSMutation contains a file system mutation: type and details.
type FSMutation struct {
	Type    MutationType
	Details EntryMutations
}

// EntryMutations contains the details of the mutation of an entry.
// There will be only one entry in the case of a Creation or a Deletion.
// There will be 2 entries in the case of an Update or a Modification, in which case the first
// entry will be the "from" state and the second will be the "to" state.
type EntryMutations []VersionedEntry

type VersionedEntry struct {
	Version
	FSEntry
}

type FSMutations []FSMutation

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
	// nolint: rowserrcheck
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT scm.mutation_type,
				fs.type,
				fs.version,
			    fs.device_id,
			    fs.entry_id,
			    fs.is_folder,
			    fs.path,
			    fs.name,
			    fs.parent_folder_id,
			    fs.created,
			    fs.modified,
			    fs.size,
			    fs.hash
		 FROM staging_cross_mutations scm
			  LEFT OUTER JOIN filesystem fs
			  ON scm.fs_name = fs.type
			     AND scm.device_id = fs.device_id
				 AND scm.entry_id = fs.entry_id
		 ORDER BY scm.mutation_type, fs.entry_id, fs.version DESC`, // `fs.version DESC`: `Previous` before `New`
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	return processFSMutationsRows(rows)
}

func processFSMutationsRows(rows *sql.Rows) (FSMutations, error) {
	fsMutations := FSMutations{}
	fsm := FSMutation{}
	previousEntryKey := ""

	for rows.Next() {
		mType, version, fsEntry, err := getMutationDetails(rows)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		newEntryKey := fmt.Sprintf("%s%s%d", fsEntry.FSName, fsEntry.DeviceID, fsEntry.EntryID)
		if newEntryKey != previousEntryKey {
			if err = sanitiseFSMDetails(fsm); err != nil {
				return nil, errors.WithMessagef(err, "FSName: '%s' DeviceID: '%s' EntryID: '%d'", fsEntry.FSName, fsEntry.DeviceID, fsEntry.EntryID)
			}

			fsMutations = append(fsMutations, fsm)
			fsm = FSMutation{
				Type: *mType,
			}
		} else {
			// by inference, at this point we are looking at the 2nd or greater element of fsm.Details
			if *mType != fsm.Type {
				return nil, errors.Errorf("both mutation details are for different types of mutation '%s' vs '%s' - FSName: '%s' DeviceID: '%s' EntryID: '%d'", *mType, fsm.Type, fsEntry.FSName, fsEntry.DeviceID, fsEntry.EntryID)
			}
		}

		ve := VersionedEntry{
			Version: *version,
			FSEntry: *fsEntry,
		}

		fsm.Details = append(fsm.Details, ve)
		previousEntryKey = newEntryKey
	}

	err := rows.Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fsMutations = append(fsMutations, fsm)

	return fsMutations[1:], nil
}

func sanitiseFSMDetails(fsm FSMutation) error {
	switch fsm.Type {
	case MutationTypeCreated, MutationTypeDeleted:
		if len(fsm.Details) != 1 {
			return errors.Errorf("mutation with more than the expected single state in Details")
		}
	case MutationTypeModified, MutationTypeMoved:
		if len(fsm.Details) != 2 {
			return errors.Errorf("mutation with more than the expected two states in Details")
		}
		if fsm.Details[0].Version == fsm.Details[1].Version {
			return errors.Errorf("both mutation details are unexpectedly for version '%s'", fsm.Details[0].Version)
		}
		if fsm.Details[0].Version == VersionNew {
			fsm.Details[0].Version, fsm.Details[1].Version = fsm.Details[1].Version, fsm.Details[0].Version
		}
	}

	return nil
}

func getMutationDetails(rows *sql.Rows) (*MutationType, *Version, *FSEntry, error) {
	var (
		mType   MutationType
		version Version
		fsEntry FSEntry
	)

	err := rows.Scan(
		&mType,
		&fsEntry.FSName,
		&version,
		&fsEntry.DeviceID,
		&fsEntry.EntryID,
		&fsEntry.IsFolder,
		&fsEntry.Path,
		&fsEntry.Name,
		&fsEntry.ParentFolderID,
		&fsEntry.Created,
		&fsEntry.Modified,
		&fsEntry.Size,
		&fsEntry.Hash,
	)
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}

	return &mType, &version, &fsEntry, nil
}

// GetFileSystemMutations returns a slice of mutations for the specified file system.
// It should be noted that up to two rows may be created: one for each version: previous and new.
func (s *SQLite3) GetFileSystemMutations(ctx context.Context, fsName FSName) (FSMutations, error) {
	// nolint: rowserrcheck
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT scm.mutation_type,
			    fs.fs_name,
			    fs.version,
			    fs.device_id,
			    fs.entry_id,
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
			  ON scm.fs_name = fs.fs_name
			  	 AND scm.device_id = fs.device_id
			  	 AND scm.entry_id = fs.entry_id
		 WHERE scm.fs_name = :fs_name
		 ORDER BY scm.mutation_type, fs.entry_id, fs.version DESC`, // `fs.version DESC`: `Previous` before `New`
		sql.Named("fs_name", fsName),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	return processFSMutationsRows(rows)
}

type SyncPair struct {
	pairName PairName
	fromFS   FSName
	toFS     FSName
}

type PairName string

func (s *SQLite3) findSyncPair(ctx context.Context, pairName PairName) (*SyncPair, error) {
	var syncPair SyncPair

	err := s.db.
		QueryRowContext(
			ctx,
			`SELECT pair_name, from_fs, to_fs
			 FROM sync_pairs
			 WHERE pair_name == :pair_name`,
			sql.Named("pair_name", pairName)).
		Scan(&syncPair.pairName, &syncPair.fromFS, &syncPair.toFS)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &syncPair, nil
}

func (s *SQLite3) findSyncPairsByFSName(ctx context.Context, fsName FSName) ([]SyncPair, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT pair_name, from_fs, to_fs
		 FROM sync_pairs
		 WHERE from_fs == :fs_name OR to_fs == :fs_name`,
		sql.Named("fs_name", fsName),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	syncPairs := []SyncPair{}

	for rows.Next() {
		sp := SyncPair{}

		err := rows.Scan(
			&sp.fromFS,
			&sp.toFS,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		syncPairs = append(syncPairs, sp)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return syncPairs, nil
}

// TODO: for this to work reliably, path needs to be standardised (i.e. \ or /, C: or /, etc).
func (s *SQLite3) refreshCrossFSMutationsStagingTable(ctx context.Context, fsName FSName) error {
	syncPairs, err := s.findSyncPairsByFSName(ctx, fsName)
	if err != nil {
		return err
	}

	for _, pair := range syncPairs {
		_, err = s.db.ExecContext(
			ctx,
			`WITH from_fs AS (SELECT type, device_id, entry_id, path, name, parent_folder_id, hash
			                FROM filesystem
						   WHERE version = :version_new
					         AND type = :from_fs),
			 to_fs AS (SELECT type, device_id, entry_id, path, name, parent_folder_id, hash
				            FROM filesystem
						   WHERE version = :version_new
						     AND type = :to_fs)

		INSERT INTO staging_cross_mutations

		SELECT
			:mutation_type_deleted,
			from_fs.type,
			from_fs.device_id,
			from_fs.entry_id
		 FROM from_fs LEFT OUTER JOIN to_fs USING (path, name)
		 WHERE to_fs.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_created,
			to_fs.type,
			to_fs.device_id,
			to_fs.entry_id
		 FROM to_fs LEFT OUTER JOIN from_fs USING (path, name)
		 WHERE from_fs.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_modified,
			to_fs.type,
			to_fs.device_id,
			to_fs.entry_id
		 FROM from_fs JOIN to_fs USING (path, name)
		 WHERE from_fs.parent_folder_id = to_fs.parent_folder_id
		   AND (
		       -- hash is not relevant for folders and that's just fine
		       from_fs.hash != to_fs.hash
		   )

		 UNION

		 SELECT
			:mutation_type_moved,
			to_fs.type,
			to_fs.device_id,
			to_fs.entry_id
		  FROM from_fs JOIN to_fs USING (path, name)
		 -- it should be noted that a file may both move and change
		 WHERE from_fs.parent_folder_id != to_fs.parent_folder_id`,
			sql.Named("from_fs", pair.fromFS),
			sql.Named("to_fs", pair.toFS),
			sql.Named("version_new", VersionNew),
			sql.Named("mutation_type_deleted", MutationTypeDeleted),
			sql.Named("mutation_type_created", MutationTypeCreated),
			sql.Named("mutation_type_modified", MutationTypeModified),
			sql.Named("mutation_type_moved", MutationTypeMoved),
		)
	}

	return errors.WithStack(err)
}

func (s *SQLite3) refreshFSMutationsStagingTable(ctx context.Context, fsName FSName) error {
	_, err := s.db.ExecContext(
		ctx,
		`WITH previous AS (SELECT fs_name, version, device_id, entry_id, parent_folder_id, hash
						     FROM filesystem
						    WHERE version = :version_previous
					          AND fs_name = :fs_name),
				  new AS (SELECT fs_name, version, device_id, entry_id, parent_folder_id, hash
						    FROM filesystem
						   WHERE version = :version_new
						     AND fs_name = :fs_name)

		INSERT INTO staging_fs_mutations

		SELECT
			:mutation_type_deleted,
			previous.fs_name,
			previous.version,
			previous.device_id,
			previous.entry_id
		 FROM previous LEFT OUTER JOIN new USING (device_id, entry_id)
		 WHERE new.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_created,
			new.fs_name,
			new.version,
			new.device_id,
			new.entry_id
		 FROM new LEFT OUTER JOIN previous USING (device_id, entry_id)
		 WHERE previous.entry_id IS NULL

		 UNION

		 SELECT
			:mutation_type_modified,
			new.fs_name,
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
			new.fs_name,
			new.version,
			new.device_id,
			new.entry_id
		  FROM new JOIN previous USING (device_id, entry_id)
		 -- it should be noted that a file may both move and change
		 WHERE new.parent_folder_id != previous.parent_folder_id`,
		sql.Named("fs_name", fsName),
		sql.Named("version_previous", VersionPrevious),
		sql.Named("version_new", VersionNew),
		sql.Named("mutation_type_deleted", MutationTypeDeleted),
		sql.Named("mutation_type_created", MutationTypeCreated),
		sql.Named("mutation_type_modified", MutationTypeModified),
		sql.Named("mutation_type_moved", MutationTypeMoved),
	)

	return errors.WithStack(err)
}

// FSName is a descriptive name for the tracked file system.
type FSName string

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
func (s *SQLite3) DeleteVersionNew(ctx context.Context, fsName FSName) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.deleteVersion(ctx, tx, fsName, VersionNew)
	if err != nil {
		return doRollback(tx, err)
	}

	err = s.deleteFSMutations(ctx, tx, fsName)
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

func (s *SQLite3) deleteVersionPrevious(ctx context.Context, tx *sql.Tx, fsName FSName) error {
	err := s.deleteVersion(ctx, tx, fsName, VersionPrevious)
	if err != nil {
		return doRollback(tx, err)
	}

	err = s.deleteFSMutations(ctx, tx, fsName)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLite3) deleteVersion(ctx context.Context, tx *sql.Tx, fsName FSName, version Version) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM "filesystem"
		 WHERE version = ?
		   AND fs_name = ?`,
		version,
		fsName,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLite3) deleteFSMutations(ctx context.Context, tx *sql.Tx, fsName FSName) error {
	_, err := tx.ExecContext(
		ctx,
		`DELETE FROM "staging_fs_mutations"
		 WHERE fs_name = ?`,
		fsName,
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

// RotateFileSystemVersions clears the "previous" file system entries for the specified
// file system and marks the "new" file system entries as "previous".
func (s *SQLite3) RotateFileSystemVersions(ctx context.Context, fsName FSName) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.deleteVersionPrevious(ctx, tx, fsName)
	if err != nil {
		return doRollback(tx, err)
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE "filesystem"
		 SET version = ?
		 WHERE version = ?
		   AND fs_name = ?`,
		VersionPrevious,
		VersionNew,
		fsName,
	)
	if err != nil {
		return doRollback(tx, err)
	}

	err = s.deleteFSMutations(ctx, tx, fsName)
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

func (s *SQLite3) FindSyncPeers(ctx context.Context, fsName FSName) ([]FSName, error) {
	panic("add a new table 'fs_status' (fs_name, fs_status) where fs_status can be ['changed', 'up-to-date']")
	panic("when the above has been done, FindSyncPeers will likely work differently and so will the other *Sync*() methods")

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT from_fs AS fs_name,
		 FROM sync_pairs
		 WHERE to_fs == :fs_name

		 UNION

		 SELECT to_fs AS fs_name,
		 FROM sync_pairs
		 WHERE from_fs == :fs_name`,
		sql.Named("fs_name", fsName),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rows.Close() }()

	peers := []FSName{}

	for rows.Next() {
		var peer FSName

		err := rows.Scan(&peer)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		peers = append(peers, peer)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return peers, nil
}

// MarkFileSystemAsChanged marks the status of the file system as "changed".
// This also triggers the internal refresh of all staging tables.
func (s *SQLite3) MarkFileSystemAsChanged(ctx context.Context, fsName FSName) error {
	panic("this should be for a sync pair, not file system by name")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	syncPeers, err := s.FindSyncPeers(ctx, fsName)
	if err != nil {
		return err
	}

	syncPeers = append(syncPeers, fsName)

	for _, fsName := range syncPeers {
		err = s.refreshFSMutationsStagingTable(ctx, fsName)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = s.refreshCrossFSMutationsStagingTable(ctx, fsName)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO "sync" ("fs_name", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("fs_name")
		 DO UPDATE SET "status" = excluded.status`,
		fsName,
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
func (s *SQLite3) MarkSyncInProgress(ctx context.Context, fsName FSName) error {
	panic("this should be for a sync pair, not file system by name")
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO "sync" ("fs_name", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("fs_name")
		 DO UPDATE SET "status" = excluded.status`,
		fsName,
		SyncStatusInProgress,
	)

	return errors.WithStack(err)
}

// MarkSyncComplete marks the status of the sync as "complete".
// TODO: it may be that the staging table should be cleared down, although not essential because
// that is properly taken care of by other methods that change the state of table "filesystem".
func (s *SQLite3) MarkSyncComplete(ctx context.Context, fsName FSName) error {
	panic("this should be for a sync pair, not file system by name")
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO "sync" ("fs_name", "status")
		 VALUES (?, ?)
		 ON CONFLICT ("fs_name")
		 DO UPDATE SET "status" = excluded.status`,
		fsName,
		SyncStatusComplete,
	)

	return errors.WithStack(err)
}

type FSInfo struct {
	FSName    FSName
	FSDriver  FSDriver
	FSRoot    string
	FSChanged bool
}

// GetFileSystemInfo returns high level information about the file system fsName.
// It will return an error (no rows found) if it cannot find the status row.
func (s *SQLite3) GetFileSystemInfo(ctx context.Context, fsName FSName) (*FSInfo, error) {
	var fsInfo FSInfo

	err := s.db.QueryRowContext(
		ctx,
		`SELECT fs_name, fs_driver, fs_root, fs_changed
		 FROM "fs_info"
		 WHERE "fs_name" = ?`,
		fsName,
	).Scan(
		&fsInfo.FSName,
		&fsInfo.FSDriver,
		&fsInfo.FSRoot,
		&fsInfo.FSChanged,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &fsInfo, nil
}

// IsFileSystemEmpty returns true if no entry data exists at all in the database for `fsName`,
// otherwise it returns false.
func (s *SQLite3) IsFileSystemEmpty(ctx context.Context, fsName FSName) (bool, error) {
	previousRowsCount := 0

	err := s.db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM "filesystem"
		 WHERE "fs_name" = ?
		   AND "version" = ?`,
		fsName,
		VersionPrevious,
	).Scan(&previousRowsCount)
	if err != nil {
		return false, errors.WithStack(err)
	}

	newRowsCount := 0

	err = s.db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM "filesystem"
		 WHERE "fs_name" = ?
		   AND "version" = ?`,
		fsName,
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
