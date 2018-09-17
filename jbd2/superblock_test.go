package jbd2

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-ext4"
	"github.com/dsoprea/go-logging"
)

func TestNewJournalSuperblock(t *testing.T) {
	filepath := path.Join(assetsPath, "journal.ext4")

	// Load the filesystem.

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(ext4.Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := ext4.NewSuperblockWithReader(f)
	log.PanicIf(err)

	if sb.HasCompatibleFeature(ext4.SbFeatureCompatHasJournal) == false {
		t.Fatalf("filesystem does not have a journal.")
	}

	inodeNumber := int(sb.Data().SJournalInum)

	if inodeNumber != ext4.InodeJournal {
		t.Fatalf("inode number different than expected: (%d) != (%d)", inodeNumber, ext4.InodeJournal)
	}

	bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err := ext4.NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	// Read the journal data.

	en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)
	ir := ext4.NewInodeReader(en)

	jsb, err := NewJournalSuperblock(ir)
	log.PanicIf(err)

	jb, err := jsb.NextBlock(ir)
	log.PanicIf(err)

	jdb := jb.(*JournalDescriptorBlock)

	if len(jdb.Tags) != 1 {
		t.Fatalf("exactly one tag was not found: (%d)", len(jdb.Tags))
	}

	tag := jdb.Tags[0]

	if tag.String() != "JournalBlockTag<TBLOCKNR=(74) TCHECKSUM=(0) TFLAGS=(8) UUID=[00000000000000000000000000000000]>" {
		t.Fatalf("descriptor tag not as expected: [%s]", tag)
	}

	_, err = jsb.NextBlock(ir)
	if err == nil {
		t.Fatalf("expected EOF for no more journal blocks")
	} else if err != io.EOF {
		log.Panic(err)
	}
}
