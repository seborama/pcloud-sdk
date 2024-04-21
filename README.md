# pCloud Go client, for the rest of us

This is a pCloud client written in Go for cross-platform compatibility, such as a Raspberry Pi in my use-case.

NOTE: I'm **not** affiliated to pCloud so this project is as good or as bad as it gets.

## Go SDK 🤩

See [SDK](sdk/README.md).

## FUSE drive for pCloud (Linux and FreeBSD) 🤩😍

See [fuse](fuse/README.md)

## Tracker (file system mutations)

See [Tracker](tracker/README.md).

## Sync (file system synchronisation)

See [Sync](sync/README.md).

## History

The original driver for this project is to create a pCloud client for my Raspberry Pi4.

While [pCloud's console client](https://github.com/pcloudcom/console-client) seemed to tick the boxes, I wasn't able to use it for two reasons:
- it creates a virtual drive - files are not stored on my local storage as I require. In my set-up, cloud storage is a convenience (backup and remote access), not an extension / replacement of my local storage.
- I elected to create a mirror from the pCloud virtual drive (that the [console-client](https://github.com/pcloudcom/console-client) creates) to my local storage. While this would work well for my needs (using pCloud as the primary source of truth, and the mirror as a local replica), I found that both `rsync` and `unison` were confronted to an I/O deadlock when downloading data. The [console-client](https://github.com/pcloudcom/console-client) appears to have an issue in its internal cache management and that blocks all I/O. Recovery involves restarting the pCloud console-client daemon but the story repeats again and the mirror cannot complete.

## Objectives

1. ✅ implement a Go version of the SDK.

2. 🧑‍💻 FUSE integration (Linux / FreeBSD)

3. implement a sync command.

4. CLI for basic pCloud interactions (copy, move, etc)

## pCloud API documentation

The official pCloud API is documented at:

https://docs.pcloud.com
