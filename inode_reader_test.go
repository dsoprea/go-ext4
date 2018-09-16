package ext4

import (
	"bytes"
	"io"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

func TestInodeReader_Read(t *testing.T) {
	f, inode, err := GetTestFileInode()
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
