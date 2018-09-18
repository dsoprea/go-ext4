package jbd2

import (
	"io"
	"os"
	"path"

	"github.com/dsoprea/go-ext4"
	"github.com/dsoprea/go-logging"
)

var (
	assetsPath = path.Join(os.Getenv("GOPATH"), "src", "github.com", "dsoprea", "go-ext4", "jbd2", "assets")
)

// GetJournalInode returns inode and file structs. It's the responsibility of
// the caller to close it.
func GetJournalInode(filepath string) (f *os.File, inode *ext4.Inode, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// Load the filesystem.

	f, err = os.Open(filepath)
	log.PanicIf(err)

	_, err = f.Seek(ext4.Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := ext4.NewSuperblockWithReader(f)
	log.PanicIf(err)

	if sb.HasCompatibleFeature(ext4.SbFeatureCompatHasJournal) == false {
		log.Panicf("filesystem does not have a journal.")
	}

	inodeNumber := int(sb.Data().SJournalInum)

	if inodeNumber != ext4.InodeJournal {
		log.Panicf("inode number different than expected: (%d) != (%d)", inodeNumber, ext4.InodeJournal)
	}

	bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err = ext4.NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	return f, inode, nil
}
