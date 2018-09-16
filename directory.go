package ext4

import (
	"fmt"

	"github.com/dsoprea/go-logging"
)

const (
	Ext4FilenameMaxLen     = 255
	Ext4DirectoryEntrySize = Ext4FilenameMaxLen + 8
)

// File types.
const (
	FileTypeUnknown         = uint8(0x0)
	FileTypeRegular         = uint8(0x1)
	FileTypeDirectory       = uint8(0x2)
	FileTypeCharacterDevice = uint8(0x3)
	FileTypeBlockDevice     = uint8(0x4)
	FileTypeFifo            = uint8(0x5)
	FileTypeSocket          = uint8(0x6)
	FileTypeSymbolicLink    = uint8(0x7)
)

var (
	FileTypeLookup = map[uint8]string{
		FileTypeUnknown:         "unknown",
		FileTypeRegular:         "regular",
		FileTypeDirectory:       "directory",
		FileTypeCharacterDevice: "character device",
		FileTypeBlockDevice:     "block device",
		FileTypeFifo:            "fifo",
		FileTypeSocket:          "socket",
		FileTypeSymbolicLink:    "symbolic link",
	}
)

// Ext4DirEntry2 is one of potentially many sequential entries stored in a
// directory inode.
type Ext4DirEntry2 struct {
	Inode    uint32 // Number of the inode that this directory entry points to.
	RecLen   uint16 // Length of this directory entry.
	NameLen  uint8  // Length of the file name.
	FileType uint8  // File type code, see ftype table below.
	Name     []byte // File name. Has a maximum size of Ext4FilenameMaxLen but actual length derived from `RecLen`.
}

// DirectoryEntry wraps the raw directory entry and provides higher-level
// functionality.
type DirectoryEntry struct {
	data *Ext4DirEntry2
}

func (de *DirectoryEntry) Data() *Ext4DirEntry2 {
	return de.data
}

func (de *DirectoryEntry) Name() string {
	return string(de.data.Name[:])
}

func (de *DirectoryEntry) IsUnknownType() bool {
	return de.data.FileType == FileTypeUnknown
}

func (de *DirectoryEntry) IsRegular() bool {
	return de.data.FileType == FileTypeRegular
}

func (de *DirectoryEntry) IsDirectory() bool {
	return de.data.FileType == FileTypeDirectory
}

func (de *DirectoryEntry) IsCharacterDevice() bool {
	return de.data.FileType == FileTypeCharacterDevice
}

func (de *DirectoryEntry) IsBlockDevice() bool {
	return de.data.FileType == FileTypeBlockDevice
}

func (de *DirectoryEntry) IsFifo() bool {
	return de.data.FileType == FileTypeFifo
}

func (de *DirectoryEntry) IsSocket() bool {
	return de.data.FileType == FileTypeSocket
}

func (de *DirectoryEntry) IsSymbolicLink() bool {
	return de.data.FileType == FileTypeSymbolicLink
}

func (de *DirectoryEntry) TypeName() string {
	name, found := FileTypeLookup[de.data.FileType]
	if found == false {
		log.Panicf("invalid type (%d) for inode (%d)", de.data.Inode)
	}

	return name
}

func (de *DirectoryEntry) String() string {
	return fmt.Sprintf("DirectoryEntry<NAME=[%s] INODE=(%d) TYPE=[%s]-(%d)>", de.Name(), de.data.Inode, de.TypeName(), de.data.FileType)
}
