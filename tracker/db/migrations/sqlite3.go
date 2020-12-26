package migrations

var SQLite3 = []string{
	`
		BEGIN;

		CREATE TABLE IF NOT EXISTS "filesystem" (
			"type"              VARCHAR,
			"version"           CHAR(1),
			"device_id"         VARCHAR,
			"entry_id"          VARCHAR,
			"is_folder"         BOOL DEFAULT FALSE,
			"is_deleted"        BOOL DEFAULT FALSE,
			"deleted_file_id"   VARCHAR NULL,
			"path"              VARCHAR NOT NULL,
			"name"              VARCHAR NOT NULL,
			"parent_folder_id"  VARCHAR NOT NULL,
			"created"           DATETIME NOT NULL,
			"modified"          DATETIME NOT NULL,
			"size"              INTEGER NULL, -- only valid for files
			"hash"              VARCHAR NULL, -- only valid for files

			PRIMARY KEY (type, version, device_id, entry_id)
		);

		CREATE INDEX IF NOT EXISTS device_entry ON filesystem (device_id, entry_id);

		COMMIT;`,
}
