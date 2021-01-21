package migrations

// SQLite3 holds the migrations for the sqlite3-based schema.
// TODO: missing a concept of replica_id which would uniquely identify a cloud or a local
//       replica beyond the current type + device_id which is not sufficient. For instance,
//       there could be 2 local replicas with the same device ID (they would differ by root
//       path).
var SQLite3 = []string{
	`	BEGIN;

		CREATE TABLE IF NOT EXISTS "sync" (
			"type"       VARCHAR,
			"status"     VARCHAR,

			PRIMARY KEY ("type")
		);

		CREATE TABLE IF NOT EXISTS "filesystem" (
			"type"              VARCHAR,
			"version"           VARCHAR,
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,
			"is_folder"         BOOL DEFAULT FALSE,
			"path"              VARCHAR NOT NULL,
			"name"              VARCHAR NOT NULL,
			"parent_folder_id"  VARCHAR NOT NULL,
			"created"           DATETIME NOT NULL,
			"modified"          DATETIME NOT NULL,
			"size"              INTEGER NULL, -- only valid for files
			"hash"              VARCHAR NULL, -- only valid for files

			PRIMARY KEY (type, version, device_id, entry_id)
		);

		CREATE INDEX IF NOT EXISTS filesystem_device_entry ON filesystem (device_id, entry_id);
		CREATE INDEX IF NOT EXISTS filesystem_device_entry ON filesystem (path, name);

		CREATE TABLE IF NOT EXISTS "staging_cross_mutations" (
			"mutation_type"     VARCHAR,
			"fs_type"           VARCHAR,
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,
			-- "is_folder"         BOOL DEFAULT FALSE,
			-- "path"              VARCHAR NOT NULL,
			-- "name"              VARCHAR NOT NULL,
			-- "parent_folder_id"  VARCHAR NOT NULL,
			-- "created"           DATETIME NOT NULL,
			-- "modified"          DATETIME NOT NULL,
			-- "size"              INTEGER NULL, -- only valid for files
			-- "hash"              VARCHAR NULL, -- only valid for files

			-- a file can both mutate and move
			PRIMARY KEY (mutation_type, fs_type, device_id, entry_id)
		);

		CREATE TABLE IF NOT EXISTS "staging_fs_mutations" (
			"mutation_type"     VARCHAR,
			"fs_type"           VARCHAR,
			"version"           VARCHAR,
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,
			-- "is_folder"         BOOL DEFAULT FALSE,
			-- "path"              VARCHAR NOT NULL,
			-- "name"              VARCHAR NOT NULL,
			-- "parent_folder_id"  VARCHAR NOT NULL,
			-- "created"           DATETIME NOT NULL,
			-- "modified"          DATETIME NOT NULL,
			-- "size"              INTEGER NULL, -- only valid for files
			-- "hash"              VARCHAR NULL, -- only valid for files

			-- a file can both mutate and move
			PRIMARY KEY (mutation_type, fs_type, version, device_id, entry_id)
		);

		CREATE INDEX IF NOT EXISTS staging_fs_mutations_fstype_version_device_entry ON staging_fs_mutations (fs_type, version, device_id, entry_id);

		COMMIT;`,
}
