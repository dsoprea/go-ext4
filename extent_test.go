package ext4

import (
	"bytes"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

func TestExtentNavigator_Block(t *testing.T) {
	f, inode, err := GetTestFileInode()
	log.PanicIf(err)

	defer f.Close()

	en := NewExtentNavigatorWithReadSeeker(f, inode)

	inodeSize := inode.Size()
	actualBytes := make([]byte, inodeSize)

	for offset := uint64(0); offset < inodeSize; {
		data, err := en.Read(offset)
		log.PanicIf(err)

		copy(actualBytes[offset:], data)
		offset += uint64(len(data))
	}

	expectedBytes, err := ioutil.ReadFile("assets/thejungle.txt")
	log.PanicIf(err)

	if bytes.Compare(actualBytes, expectedBytes) != 0 {
		t.Fatalf("Bytes not read correctly.")
	}
}
