# pCloud Go client, for the rest of us

This is a pCloud client written in Go for cross-platform compatibility.

The first objective is to implement a Go version of the SDK.

NOTE: I'm not affiliated to pCloud so this project is as good or as bad as it gets.

## Documentation

The pCloud SDK is documented at:

https://docs.pcloud.com

## Getting started

The tests rely on the presence of two environment variables to supply your credentials:
- `GO_PCLOUD_USERNAME`
- `GO_PCLOUD_PASSWORD`

At the moment, tests also require the presence of a folder named `Test` in the root of your pCloud drive, with the following contents:

```
/Test
├── Getting\ started\ with\ pCloud.pdf
└── My\ Folder
    ├── File 1.pdf
    └── My Inner Folder
        ├── File2.pdf
        └── File3.pdf
```

The actual contents of the files is currently of no relevance.
