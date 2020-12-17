# pCloud Go client, for the rest of us

This is a pCloud client written in Go for cross-platform compatibility, such as a Raspberry Pi in my use-case.

NOTE: I'm **not** affiliated to pCloud so this project is as good or as bad as it gets.

## History

The original driver for this project is to create a pCloud client for my Raspberry Pi4.

While [pCloud's console client](https://github.com/pcloudcom/console-client) seemed to tick the boxes, I wasn't able to use it for two reasons:
- it creates a virtual drive - files are not stored on my local storage as I require. In my set-up, cloud storage is a convenience (backup and remote access), not an extension / replacement of my local storage.
- I elected to create a mirror from the pCloud virtual drive (that the [console-client](https://github.com/pcloudcom/console-client) creates) to my local storage. While this would work well for my needs (using the pCloud as the primary source of truth, and the mirror as a local replica), I found that both `rsync` and `unison` hit an I/O deadlock when downloading data. The [console-client](https://github.com/pcloudcom/console-client) seems to hit a problem with its internal cache management and blocks all I/O. Recovery involves restarting the pCloud console-client daemon but the story repeats again and the mirror cannot complete.

The first objective is to implement a Go version of the SDK.

## Documentation

The official pCloud SDK is documented at:

https://docs.pcloud.com

## Getting started

The tests rely on the presence of two environment variables to supply your credentials:
- `GO_PCLOUD_USERNAME`
- `GO_PCLOUD_PASSWORD`
