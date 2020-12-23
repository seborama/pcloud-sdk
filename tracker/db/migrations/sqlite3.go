package migrations

var SQLite3 = []string{
	`
		BEGIN;

		CREATE TABLE IF NOT EXISTS "tracker_diff" (
			"diff_id"    INTEGER PRIMARY KEY,
			"timestamp"  DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS "filesystem" (
			"version"           CHAR(1),
			"entry_id"          VARCHAR,
			"is_folder"         BOOL DEFAULT FALSE,
			"is_deleted"        BOOL DEFAULT FALSE,
			"deleted_file_id"   VARCHAR NULL,
			"name"              VARCHAR NOT NULL,
			"parent_folder_id"  VARCHAR NOT NULL,
			"created"           DATETIME NOT NULL,
			"modified"          DATETIME NOT NULL,
			"size"              INTEGER NULL, -- only valid for files
			"hash"              VARCHAR NULL, -- only valid for files

			PRIMARY KEY (version, entry_id)
		);

		CREATE TABLE IF NOT EXISTS "events" (
			"diff_id"    INTEGER,
			"file_id"    VARCHAR,
			"created"    DATETIME NOT NULL,
			PRIMARY KEY (diff_id, file_id)
		);

		COMMIT;`,
}
