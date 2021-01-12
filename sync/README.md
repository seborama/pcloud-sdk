# Sync

## Getting started

```bash
make build
make test-sync
```

## Status

- TBC Supports local file systems for Linux and OSX (Windows??).

## Noteworthy

- left sync and right sync are evetually consistent by nature: if the source changes while the sync is running, the next sync will re-align the destination.
- two-way sync is more problematic by nature AND NO COMPLETELY SAFE SOLUTION IS AVAILABLE YET. Example: if side A changes during the sync, side B will not have the correct state. On subsequent sync, if side B was modifed, it would be synced to side A although it did not have the right state of sync. As a minimum, this must result in a conflict that should be resolved manually. Other options:
    - locking files at the start of sync and release them when after each file has been synced (limitation: folders can't be locked)
    - file system notifications (pCloud has support but other cloud providers may not)
    - **check hash of source files before and after sync and rollback destination if the hash has changed. For folders, check the source folder has not moved after sync (this requires checking the folder's EntryID too so to guarantee it is the same folder), and rollback if changed**
