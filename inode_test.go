package ext4

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestNewInodeWithReadSeeker_RootInode(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	inodeNumber := 2
	// inodeNumber := 12

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err := NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	actualTimestamp := inode.InodeChangeTime().String()
	if actualTimestamp != "2018-09-08 02:08:45 -0400 EDT" {
		t.Fatalf("InodeChangeTime() timestamp not correct: [%s]", actualTimestamp)
	}
}

func TestNewInodeWithReadSeeker_FileInode(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	inodeNumber := 12

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err := NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	actualTimestamp := inode.InodeChangeTime().String()
	if actualTimestamp != "2018-09-08 02:08:45 -0400 EDT" {
		t.Fatalf("InodeChangeTime() timestamp not correct: [%s]", actualTimestamp)
	}

	// TODO(dustin): !! For experimentation/debugging.
	// inode.DumpDirectory()
}