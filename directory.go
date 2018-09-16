package ext4

import (
	"bytes"
	"fmt"
	"io"

	"encoding/binary"

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
	return fmt.Sprintf("DirectoryEntry<NAME=[%s] NAME-LEN=(%d)/(%d) INODE=(%d) TYPE=[%s]-(%d)>", de.Name(), len(de.Name()), de.data.NameLen, de.data.Inode, de.TypeName(), de.data.FileType)
}

// DirectoryBrowser provides high-level directory navigation.
type DirectoryBrowser struct {
	inodeReader *InodeReader

	dataSize uint64
	dataRead uint64
}

func NewDirectoryBrowser(rs io.ReadSeeker, inode *Inode) *DirectoryBrowser {
	en := NewExtentNavigatorWithReadSeeker(rs, inode)
	ir := NewInodeReader(en)

	return &DirectoryBrowser{
		inodeReader: ir,
		dataSize:    inode.Size(),
	}
}

// Next parses the next directory entry from the underlying inode data reader.
// Returns `io.EOF` when done.
func (db *DirectoryBrowser) Next() (de *DirectoryEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// TODO(dustin): !! We should be seeing an extra `ext4_dir_entry_tail` as the last entry, which has a similar structure to a normal entry but provides a checksum. However, it doesn't appear to be present.

	if db.dataRead >= db.dataSize {
		return nil, io.EOF
	}

	raw := new(Ext4DirEntry2)

	err = binary.Read(db.inodeReader, binary.LittleEndian, &raw.Inode)
	log.PanicIf(err)

	err = binary.Read(db.inodeReader, binary.LittleEndian, &raw.RecLen)
	log.PanicIf(err)

	// Read the remaining data, which is variable-length.

	offset := 0
	needBytes := int(raw.RecLen) - 6
	record := make([]byte, needBytes)

	for offset < needBytes {
		n, err := db.inodeReader.Read(record[offset:])
		log.PanicIf(err)

		offset += n
	}

	b := bytes.NewBuffer(record)

	err = binary.Read(b, binary.LittleEndian, &raw.NameLen)
	log.PanicIf(err)

	err = binary.Read(b, binary.LittleEndian, &raw.FileType)
	log.PanicIf(err)

	raw.Name = record[2 : 2+raw.NameLen]

	// Done. Wrap up.

	db.dataRead += uint64(raw.RecLen)

	if db.dataRead > db.dataSize {
		log.Panicf("inode size is not aligned to the directory-entry record size (we ran over): (%d) > (%d)", db.dataRead, db.dataSize)
	}

	de = &DirectoryEntry{
		data: raw,
	}

	return de, nil
}
