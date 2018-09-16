package ext4

import (
	"bytes"
	"io"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

func TestInodeReader_Read(t *testing.T) {
	f, inode, err := GetTestInode(TestFileInodeNumber)
	log.PanicIf(err)

	defer f.Close()

	en := NewExtentNavigatorWithReadSeeker(f, inode)

	var r io.Reader
	r = NewInodeReader(en)

	actualBytes, err := ioutil.ReadAll(r)
	log.PanicIf(err)

	expectedBytes, err := ioutil.ReadFile("assets/thejungle.txt")
	log.PanicIf(err)

	if bytes.Compare(actualBytes, expectedBytes) != 0 {
		t.Fatalf("Bytes not read correctly.")
	}
}

func TestInodeReader_Skip(t *testing.T) {
	f, inode, err := GetTestInode(TestFileInodeNumber)
	log.PanicIf(err)

	defer f.Close()

	en := NewExtentNavigatorWithReadSeeker(f, inode)

	ir := NewInodeReader(en)

	skipCount := uint64(1000)
	for skipCount > 0 {
		n, err := ir.Skip(skipCount)
		log.PanicIf(err)

		skipCount -= n
	}

	actualBytes, err := ioutil.ReadAll(ir)
	log.PanicIf(err)

	expectedBytes, err := ioutil.ReadFile("assets/thejungle.txt")
	log.PanicIf(err)

	if bytes.Compare(actualBytes, expectedBytes[1000:]) != 0 {
		t.Fatalf("Bytes not read correctly.")
	}
}
