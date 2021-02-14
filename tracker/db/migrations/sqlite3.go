package migrations

// SQLite3 holds the migrations for the sqlite3-based schema.
var SQLite3 = []string{
	`	BEGIN;

		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS "filesystem" (
			"fs_name"           VARCHAR,
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

			PRIMARY KEY (fs_name, version, device_id, entry_id)
		);

		CREATE INDEX IF NOT EXISTS filesystem_device_entry ON filesystem (device_id, entry_id);
		CREATE INDEX IF NOT EXISTS filesystem_device_entry ON filesystem (path, name);

		CREATE TABLE IF NOT EXISTS "fs_info" (
			"fs_name"     VARCHAR,
			"fs_driver"   VARCHAR,
			"fs_root"     VARCHAR,
			"fs_changed"  BOOL      DEFAULT FALSE,

			PRIMARY KEY ("fs_name"),

			CONSTRAINT fk_fs_name
			FOREIGN KEY(fs_name) 
			REFERENCES filesystem(fs_name)
		);

		CREATE TABLE IF NOT EXISTS "staging_cross_mutations" (
			"mutation_type"     VARCHAR,
			"fs_name"           VARCHAR,
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,

			-- a file can both mutate and move
			PRIMARY KEY (mutation_type, fs_name, device_id, entry_id)
		);

		CREATE TABLE IF NOT EXISTS "staging_fs_mutations" (
			"mutation_type"     VARCHAR,
			"fs_name"           VARCHAR,
			"version"           VARCHAR,
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,

			-- a file can both mutate and move
			PRIMARY KEY (mutation_type, fs_name, version, device_id, entry_id)
		);

		CREATE INDEX IF NOT EXISTS staging_fs_mutations_fsname_version_device_entry ON staging_fs_mutations (fs_name, version, device_id, entry_id);

		COMMIT;`,
}
