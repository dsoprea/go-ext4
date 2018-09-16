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

	entryDescriptions := make([]string, 0)

	for {
		de, err := db.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panic(err)
		}

		entryDescriptions = append(entryDescriptions, de.String())
	}

	sort.Strings(entryDescriptions)

	expectedEntryDescriptions := []string{
		"DirectoryEntry<NAME=[..] INODE=(2) TYPE=[directory]-(2)>",
		"DirectoryEntry<NAME=[.] INODE=(2) TYPE=[directory]-(2)>",
		"DirectoryEntry<NAME=[lost+found] INODE=(11) TYPE=[directory]-(2)>",
		"DirectoryEntry<NAME=[thejungle.txt] INODE=(12) TYPE=[regular]-(1)>",
	}

	if reflect.DeepEqual(entryDescriptions, expectedEntryDescriptions) == false {
		t.Fatalf("Root directory entries are not correct.")
	}
}
