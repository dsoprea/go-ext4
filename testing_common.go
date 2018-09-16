package ext4

import (
	"io"
	"os"
	"path"

	"github.com/dsoprea/go-logging"
)

const (
	TestDirectoryInodeNumber = 2
	TestFileInodeNumber      = 12
)

var (
	assetsPath = path.Join(os.Getenv("GOPATH"), "src", "github.com", "dsoprea", "go-ext4", "assets")
)

// GetTestInode returns a test inode struct and `os.File` for the file. It's
// the responsibility of the caller to close it.
func GetTestInode(inodeNumber int) (f *os.File, inode *Inode, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err = os.Open(filepath)
	log.PanicIf(err)

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err = NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	return f, inode, nil
}

func GetInode(filesystemPath string, inodeNumber int) (f *os.File, inode *Inode, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	f, err = os.Open(filesystemPath)
	log.PanicIf(err)

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err = NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	return f, inode, nil
}
