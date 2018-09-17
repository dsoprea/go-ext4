[![Build Status](https://travis-ci.org/dsoprea/go-ext4.svg?branch=master)](https://travis-ci.org/dsoprea/go-ext4)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-ext4/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-ext4?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-ext4?status.svg)](https://godoc.org/github.com/dsoprea/go-ext4)

## Overview

This package allows you to browse an *ext4* filesystem directly. It does not use FUSE or touch the kernel, so no privileges are required.


## Examples

Usage examples are [here](https://godoc.org/github.com/dsoprea/go-ext4#pkg-examples).


## Notes

- Modern filesystems are supported, including both 32-bit and 64-bit addresses (though 64-bit is not well-documented). Obscure filesystem options may not be compatible. See the [compatibility assertions](https://github.com/dsoprea/go-ext4/blob/master/superblock.go) in `NewSuperblockWithReader`.
