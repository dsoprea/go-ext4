package ext4

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

func TestExtentNavigator_Block(t *testing.T) {
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

	// inodeNumber := 2
	inodeNumber := 12

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	log.PanicIf(err)

	inode, err := NewInodeWithReadSeeker(bgd, f, inodeNumber)
	log.PanicIf(err)

	en := NewExtentNavigatorWithReadSeeker(f, inode)

	expectedBytes, err := ioutil.ReadFile("assets/thejungle.txt")
	log.PanicIf(err)

	actualBytes := make([]byte, len(expectedBytes))

	for offset := uint64(0); offset < inode.Size(); {
		data, err := en.Read(offset)
		log.PanicIf(err)

		copy(actualBytes[offset:], data)
		offset += uint64(len(data))
	}

	if bytes.Compare(actualBytes, expectedBytes) != 0 {
		t.Fatalf("Bytes not read correctly.")
	}
}
