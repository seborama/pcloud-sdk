# fuse

This package offers a pCloud client for Linux and FreeBSD for the rest of us who have been forgotten...

It uses FUSE to mount the pCloud drive. This is possible thanks to [Bazil](https://github.com/bazil) and his [FUSE library for Go](https://github.com/bazil/fuse).

I am developing on a Linux ARM Raspberry Pi4. I haven't (yet) tried Linux x86_64 or FreeBSD, it simply is too early at this stage of the development to worry about more than one platform.

## Status

At this stage, this is explorative. The code base is entirely experimental, most features are not implemented or only partially.

The drive can be mounted via the tests and it can be "walked" through.

No write operations are supported for now.

## Change log

2024-Apr-23 - Added support to read file contents.

2024-Apr-22 - The pCloud drive can listed entirely. `ls` on the root of the mount will list directories and files contained in the root of the pCloud drive.

2024-Apr-21 - The pCloud drive can be mounted (via the test - see "Getting started"). `ls` on the root of the mount will list directories and files contained in the root of the pCloud drive.

## Getting started

While this is under construction, only a simple test exists.

It mounts pCloud under `/tmp/pcloud_mnt`.

To cleanly end the test, make sure you run `umount /tmp/pcloud_mnt` on your Linux / FreeBSD command line.

Should the test end abruptly, or time out, run `umount /tmp/pcloud_mnt` to clean up the mount.

The tests rely on the presence of environment variables to supply your credentials (**make sure you `export` the variables!**):
- `GO_PCLOUD_USERNAME`
- `GO_PCLOUD_PASSWORD`
- `GO_PCLOUD_TFA_CODE` - BETA. Note that the device is automatically marked as trusted so TFA is not required the next time. You can remove the trust manually in your [account security settings](https://my.pcloud.com/#page=settings&settings=tab-security).

TFA was possible thanks to [Glib Dzevo](https://github.com/gdzevo) and his [console-client PR](https://github.com/pcloudcom/console-client/pull/94) where I found the info I needed!

```bash
cd fuse
go test -v ./

# in a separate terminal window:
ls /tmp/pcloud_mnt
# ...

# when you're done:
umount /tmp/pcloud_mnt
```
