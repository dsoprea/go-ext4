package ext4

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"io/ioutil"

	log "github.com/dsoprea/go-logging"
)

func TestExtentNavigator_Read(t *testing.T) {
	f, inode, err := GetTestInode(TestFileInodeNumber)
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

func ExampleExtentNavigator_Read() {
	f, inode, err := GetTestInode(TestFileInodeNumber)
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

	// Take the first part, strip some nonprintable characters from the front,
	// and normalize the newlines so that this example tests properly.
	firstPart := string(actualBytes[3:100])
	firstPart = strings.Replace(firstPart, "\r\n", "\n", -1)

	fmt.Println(firstPart)

	// Output:
	//
	// The Project Gutenberg EBook of The Jungle, by Upton Sinclair
	//
	// This eBook is for the use of anyo
}

func TestExtentNavigator_ReadSymlink(t *testing.T) {
	f, inode, err := GetTestInode(TestSymlinkInodeNumber)
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

	expectedBytes := []byte("thejungle.txt")

	if bytes.Compare(actualBytes, expectedBytes) != 0 {
		t.Fatalf("Bytes not read correctly.")
	}
}
