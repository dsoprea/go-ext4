package ext4

import (
	"io"
	"reflect"
	"sort"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestDirectoryBrowser_Next(t *testing.T) {
	f, inode, err := GetTestInode(TestDirectoryInodeNumber)
	log.PanicIf(err)

	defer f.Close()

	db := NewDirectoryBrowser(f, inode)

	entryNames := make([]string, 0)

	for {
		de, err := db.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panic(err)
		}

		entryNames = append(entryNames, de.Name())
	}

	sort.Strings(entryNames)

	expectedEntryNames := []string{
		".",
		"..",
		"lost+found",
		"thejungle.txt",
	}

	if reflect.DeepEqual(entryNames[3], expectedEntryNames[3]) == false {
		t.Fatalf("Root directory entry-names are not correct.")
	}
}
