package ext4

import (
	"fmt"
	"io"
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestDirectoryWalk_Next32(t *testing.T) {
	inodeNumber := InodeRootDirectory

	filepath := path.Join(assetsPath, "hierarchy_32.ext4")

	f, inode, err := GetInode(filepath, inodeNumber)
	log.PanicIf(err)

	defer f.Close()

	bgd := inode.BlockGroupDescriptor()

	dw, err := NewDirectoryWalk(f, bgd, inodeNumber)
	log.PanicIf(err)

	allEntries := make([]string, 0)

	for {
		fullPath, de, err := dw.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panic(err)
		}

		description := fmt.Sprintf("%s: %s", fullPath, de.String())
		allEntries = append(allEntries, description)
	}

	sort.Strings(allEntries)

	expectedAllEntries := []string{
		"directory1/fortune1: DirectoryEntry<NAME=[fortune1] INODE=(15) TYPE=[regular]-(1)>",
		"directory1/fortune2: DirectoryEntry<NAME=[fortune2] INODE=(14) TYPE=[regular]-(1)>",
		"directory1/fortune5: DirectoryEntry<NAME=[fortune5] INODE=(20) TYPE=[regular]-(1)>",
		"directory1/fortune6: DirectoryEntry<NAME=[fortune6] INODE=(21) TYPE=[regular]-(1)>",
		"directory1/subdirectory1/fortune3: DirectoryEntry<NAME=[fortune3] INODE=(17) TYPE=[regular]-(1)>",
		"directory1/subdirectory1/fortune4: DirectoryEntry<NAME=[fortune4] INODE=(18) TYPE=[regular]-(1)>",
		"directory1/subdirectory1: DirectoryEntry<NAME=[subdirectory1] INODE=(16) TYPE=[directory]-(2)>",
		"directory1/subdirectory2/fortune7: DirectoryEntry<NAME=[fortune7] INODE=(22) TYPE=[regular]-(1)>",
		"directory1/subdirectory2/fortune8: DirectoryEntry<NAME=[fortune8] INODE=(23) TYPE=[regular]-(1)>",
		"directory1/subdirectory2: DirectoryEntry<NAME=[subdirectory2] INODE=(19) TYPE=[directory]-(2)>",
		"directory1: DirectoryEntry<NAME=[directory1] INODE=(13) TYPE=[directory]-(2)>",
		"directory2/fortune10: DirectoryEntry<NAME=[fortune10] INODE=(26) TYPE=[regular]-(1)>",
		"directory2/fortune9: DirectoryEntry<NAME=[fortune9] INODE=(25) TYPE=[regular]-(1)>",
		"directory2: DirectoryEntry<NAME=[directory2] INODE=(24) TYPE=[directory]-(2)>",
		"lost+found: DirectoryEntry<NAME=[lost+found] INODE=(11) TYPE=[directory]-(2)>",
		"thejungle.txt: DirectoryEntry<NAME=[thejungle.txt] INODE=(12) TYPE=[regular]-(1)>",
	}

	if reflect.DeepEqual(allEntries, expectedAllEntries) == false {
		t.Fatalf("hierarchy not correct")
	}
}

func TestDirectoryWalk_Next64(t *testing.T) {
	inodeNumber := InodeRootDirectory

	filepath := path.Join(assetsPath, "hierarchy_64.ext4")

	f, inode, err := GetInode(filepath, inodeNumber)
	log.PanicIf(err)

	defer f.Close()

	bgd := inode.BlockGroupDescriptor()

	dw, err := NewDirectoryWalk(f, bgd, inodeNumber)
	log.PanicIf(err)

	allEntries := make([]string, 0)

	for {
		fullPath, de, err := dw.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panic(err)
		}

		description := fmt.Sprintf("%s: %s", fullPath, de.String())
		allEntries = append(allEntries, description)
	}

	sort.Strings(allEntries)

	expectedAllEntries := []string{
		"directory1/fortune1: DirectoryEntry<NAME=[fortune1] INODE=(17) TYPE=[regular]-(1)>",
		"directory1/fortune2: DirectoryEntry<NAME=[fortune2] INODE=(18) TYPE=[regular]-(1)>",
		"directory1/fortune5: DirectoryEntry<NAME=[fortune5] INODE=(19) TYPE=[regular]-(1)>",
		"directory1/fortune6: DirectoryEntry<NAME=[fortune6] INODE=(20) TYPE=[regular]-(1)>",
		"directory1/subdirectory1/fortune3: DirectoryEntry<NAME=[fortune3] INODE=(21) TYPE=[regular]-(1)>",
		"directory1/subdirectory1/fortune4: DirectoryEntry<NAME=[fortune4] INODE=(22) TYPE=[regular]-(1)>",
		"directory1/subdirectory1: DirectoryEntry<NAME=[subdirectory1] INODE=(14) TYPE=[directory]-(2)>",
		"directory1/subdirectory2/fortune7: DirectoryEntry<NAME=[fortune7] INODE=(23) TYPE=[regular]-(1)>",
		"directory1/subdirectory2/fortune8: DirectoryEntry<NAME=[fortune8] INODE=(24) TYPE=[regular]-(1)>",
		"directory1/subdirectory2: DirectoryEntry<NAME=[subdirectory2] INODE=(15) TYPE=[directory]-(2)>",
		"directory1: DirectoryEntry<NAME=[directory1] INODE=(12) TYPE=[directory]-(2)>",
		"directory2/fortune10: DirectoryEntry<NAME=[fortune10] INODE=(25) TYPE=[regular]-(1)>",
		"directory2/fortune9: DirectoryEntry<NAME=[fortune9] INODE=(26) TYPE=[regular]-(1)>",
		"directory2: DirectoryEntry<NAME=[directory2] INODE=(13) TYPE=[directory]-(2)>",
		"lost+found: DirectoryEntry<NAME=[lost+found] INODE=(11) TYPE=[directory]-(2)>",
		"thejungle.txt: DirectoryEntry<NAME=[thejungle.txt] INODE=(16) TYPE=[regular]-(1)>",
	}

	if reflect.DeepEqual(allEntries, expectedAllEntries) == false {
		t.Fatalf("hierarchy not correct")
	}
}
