[![Build Status](https://travis-ci.org/dsoprea/go-ext4.svg?branch=master)](https://travis-ci.org/dsoprea/go-ext4)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-ext4/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-ext4?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-ext4?status.svg)](https://godoc.org/github.com/dsoprea/go-ext4)

## Overview

This package allows you to browse an *ext4* filesystem directly. It does not use FUSE or touch the kernel, so no privileges are required.

This package also exposes the data in the journal (if one is available).


## Examples

Usage examples are [here](https://godoc.org/github.com/dsoprea/go-ext4#pkg-examples).


## Notes

- Modern filesystems are supported, including both 32-bit and 64-bit addressing. Obscure filesystem options may not be compatible. See the [compatibility assertions](https://github.com/dsoprea/go-ext4/blob/master/superblock.go) in `NewSuperblockWithReader`.
  - 64-bit addressing should be fine, as the high addressing should likely be zero when 64-bit addressing is turned-off (which is primary what our unit-tests test with). However, the available documentation is limited on the subject. It's specifically not clear which of the high/low addressing should be used or ignored when 64-bit addressing is turned-off.
