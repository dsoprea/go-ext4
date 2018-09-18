[![Build Status](https://travis-ci.org/dsoprea/go-ext4.svg?branch=master)](https://travis-ci.org/dsoprea/go-ext4)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-ext4/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-ext4?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-ext4?status.svg)](https://godoc.org/github.com/dsoprea/go-ext4)

## Overview

This package allows you to browse an *ext4* filesystem directly. It does not use FUSE or touch the kernel, so no privileges are required.

This package also exposes the data in the journal (if one is available).


## Example

Recursively walk all of the files in the filesystem:

```
inodeNumber := InodeRootDirectory

filepath := path.Join(assetsPath, "hierarchy_32.ext4")

f, err := os.Open(filepath)
log.PanicIf(err)

defer f.Close()

_, err = f.Seek(Superblock0Offset, io.SeekStart)
log.PanicIf(err)

sb, err := NewSuperblockWithReader(f)
log.PanicIf(err)

bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
log.PanicIf(err)

bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
log.PanicIf(err)

dw, err := NewDirectoryWalk(f, bgd, inodeNumber)
log.PanicIf(err)

allEntries := make([]string, 0)

for {
	fullPath, de, err := dw.Next()
	if err == io.EOF {
		break
	} else if err != nil {
		log.Panic(err)
	}

	description := fmt.Sprintf("%s: %s", fullPath, de.String())
	allEntries = append(allEntries, description)
}

sort.Strings(allEntries)

for _, entryDescription := range allEntries {
	fmt.Println(entryDescription)
}

// Output:
//
// directory1/fortune1: DirectoryEntry<NAME=[fortune1] INODE=(15) TYPE=[regular]-(1)>
// directory1/fortune2: DirectoryEntry<NAME=[fortune2] INODE=(14) TYPE=[regular]-(1)>
// directory1/fortune5: DirectoryEntry<NAME=[fortune5] INODE=(20) TYPE=[regular]-(1)>
// directory1/fortune6: DirectoryEntry<NAME=[fortune6] INODE=(21) TYPE=[regular]-(1)>
// directory1/subdirectory1/fortune3: DirectoryEntry<NAME=[fortune3] INODE=(17) TYPE=[regular]-(1)>
// directory1/subdirectory1/fortune4: DirectoryEntry<NAME=[fortune4] INODE=(18) TYPE=[regular]-(1)>
// directory1/subdirectory1: DirectoryEntry<NAME=[subdirectory1] INODE=(16) TYPE=[directory]-(2)>
// directory1/subdirectory2/fortune7: DirectoryEntry<NAME=[fortune7] INODE=(22) TYPE=[regular]-(1)>
// directory1/subdirectory2/fortune8: DirectoryEntry<NAME=[fortune8] INODE=(23) TYPE=[regular]-(1)>
// directory1/subdirectory2: DirectoryEntry<NAME=[subdirectory2] INODE=(19) TYPE=[directory]-(2)>
// directory1: DirectoryEntry<NAME=[directory1] INODE=(13) TYPE=[directory]-(2)>
// directory2/fortune10: DirectoryEntry<NAME=[fortune10] INODE=(26) TYPE=[regular]-(1)>
// directory2/fortune9: DirectoryEntry<NAME=[fortune9] INODE=(25) TYPE=[regular]-(1)>
// directory2: DirectoryEntry<NAME=[directory2] INODE=(24) TYPE=[directory]-(2)>
// lost+found: DirectoryEntry<NAME=[lost+found] INODE=(11) TYPE=[directory]-(2)>
// thejungle.txt: DirectoryEntry<NAME=[thejungle.txt] INODE=(12) TYPE=[regular]-(1)>
```

This example and others are documented [here](https://godoc.org/github.com/dsoprea/go-ext4#pkg-examples).


## Notes

- Modern filesystems are supported, including both 32-bit and 64-bit addressing. Obscure filesystem options may not be compatible. See the [compatibility assertions](https://github.com/dsoprea/go-ext4/blob/master/superblock.go) in `NewSuperblockWithReader`.
  - 64-bit addressing should be fine, as the high addressing should likely be zero when 64-bit addressing is turned-off (which is primary what our unit-tests test with). However, the available documentation is limited on the subject. It's specifically not clear which of the high/low addressing should be used or ignored when 64-bit addressing is turned-off.
