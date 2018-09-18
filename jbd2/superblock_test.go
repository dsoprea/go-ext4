package jbd2

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/dsoprea/go-ext4"
	"github.com/dsoprea/go-logging"
)

func TestNewJournalSuperblock_NextBlock(t *testing.T) {
	filepath := path.Join(assetsPath, "journal.ext4")

	f, inode, err := GetJournalInode(filepath)
	log.PanicIf(err)

	// Read the journal data.

	en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)
	ir := ext4.NewInodeReader(en)

	jsb, err := NewJournalSuperblock(ir)
	log.PanicIf(err)

	// Check first block.

	jb, err := jsb.NextBlock(ir)
	log.PanicIf(err)

	if jb.Type() != BtDescriptor {
		t.Fatalf("Expected descriptor for first block.")
	} else if jb.String() != "DescriptorBlock<TAGS=(1) DATA-LENGTH=(1024)>" {
		t.Fatalf("Descriptor not correct in first block: [%s]", jb.String())
	}

	jdb := jb.(*JournalDescriptorBlock)

	if len(jdb.Tags) != 1 {
		t.Fatalf("exactly one tag was not found: (%d)", len(jdb.Tags))
	}

	tag := jdb.Tags[0]

	if tag.String() != "JournalBlockTag32<TBLOCKNR=(74) TCHECKSUM=(0) TFLAGS=(8) UUID=[00000000000000000000000000000000]>" {
		t.Fatalf("descriptor tag not as expected: [%s]", tag.String())
	}

	// Check second block.

	jb, err = jsb.NextBlock(ir)
	log.PanicIf(err)

	if jb.Type() != BtBlockCommitRecord {
		t.Fatalf("Expected commit-block for second block.")
	} else if jb.String() != "CommitBlock<HChksumType=(0) HChksumSize=(0) CommitTime=[2018-09-18 03:34:36.58801818 +0000 UTC]>" {
		t.Fatalf("commit-block not correct in second block: [%s]", jb.String())
	}

	// Check third block.

	jb, err = jsb.NextBlock(ir)
	log.PanicIf(err)

	if jb.Type() != BtDescriptor {
		t.Fatalf("Expected descriptor for third block.")
	} else if jb.String() != "DescriptorBlock<TAGS=(6) DATA-LENGTH=(1024)>" {
		t.Fatalf("Descriptor not correct in third block: [%s]", jb.String())
	}

	// Check that there are no more blocks.

	_, err = jsb.NextBlock(ir)
	if err == nil {
		t.Fatalf("expected EOF for no more journal blocks")
	} else if err != io.EOF {
		log.Panic(err)
	}
}

func ExampleNewJournalSuperblock_NextBlock_Blocks() {
	filepath := path.Join(assetsPath, "journal.ext4")

	f, inode, err := GetJournalInode(filepath)
	log.PanicIf(err)

	en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)
	ir := ext4.NewInodeReader(en)

	jsb, err := NewJournalSuperblock(ir)
	log.PanicIf(err)

	for {
		jb, err := jsb.NextBlock(ir)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Panic(err)
		}

		fmt.Printf("%s\n", jb)
	}

	// Output:
	//
	// DescriptorBlock<TAGS=(1) DATA-LENGTH=(1024)>
	// CommitBlock<HChksumType=(0) HChksumSize=(0) CommitTime=[2018-09-18 03:34:36.58801818 +0000 UTC]>
	// DescriptorBlock<TAGS=(6) DATA-LENGTH=(1024)>
}

func ExampleNewJournalSuperblock_NextBlock_Descriptors() {
	filepath := path.Join(assetsPath, "journal.ext4")

	f, inode, err := GetJournalInode(filepath)
	log.PanicIf(err)

	en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)
	ir := ext4.NewInodeReader(en)

	jsb, err := NewJournalSuperblock(ir)
	log.PanicIf(err)

	for {
		jb, err := jsb.NextBlock(ir)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Panic(err)
		}

		if jb.Type() != BtDescriptor {
			continue
		}

		jdb := jb.(*JournalDescriptorBlock)
		jdb.Dump()
	}

	// Output:
	//
	// DescriptorBlock<TAGS=(1) DATA-LENGTH=(1024)>
	//
	//   TAG(0): JournalBlockTag32<TBLOCKNR=(74) TCHECKSUM=(0) TFLAGS=(8) UUID=[00000000000000000000000000000000]>
	//
	// DescriptorBlock<TAGS=(6) DATA-LENGTH=(1024)>
	//
	//   TAG(0): JournalBlockTag32<TBLOCKNR=(58) TCHECKSUM=(0) TFLAGS=(0) UUID=[00000000000000000000000000000000]>
	//   TAG(1): JournalBlockTag32<TBLOCKNR=(2) TCHECKSUM=(0) TFLAGS=(2) UUID=[00000000000000000000000000000000]>
	//   TAG(2): JournalBlockTag32<TBLOCKNR=(75) TCHECKSUM=(0) TFLAGS=(2) UUID=[00000000000000000000000000000000]>
	//   TAG(3): JournalBlockTag32<TBLOCKNR=(74) TCHECKSUM=(0) TFLAGS=(2) UUID=[00000000000000000000000000000000]>
	//   TAG(4): JournalBlockTag32<TBLOCKNR=(44) TCHECKSUM=(0) TFLAGS=(2) UUID=[00000000000000000000000000000000]>
	//   TAG(5): JournalBlockTag32<TBLOCKNR=(43) TCHECKSUM=(0) TFLAGS=(10) UUID=[00000000000000000000000000000000]>
}
