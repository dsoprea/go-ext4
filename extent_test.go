package ext4

import (
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

	pBlock1, err := en.PhysicalBlock(1)
	log.PanicIf(err)

	data1, err := sb.ReadPhysicalBlock(pBlock1, sb.BlockSize())
	log.PanicIf(err)

	pBlock2, err := en.PhysicalBlock(2)
	log.PanicIf(err)

	data2, err := sb.ReadPhysicalBlock(pBlock2, sb.BlockSize())
	log.PanicIf(err)

	actual := string(data1) + string(data2)

	// We need to preserve newlines and which other characters we had trouble
	// with when pasting in a raw string literal.
	expectedBytes, err := ioutil.ReadFile("assets/TestExtentNavigator_Block_expected.txt")
	log.PanicIf(err)

	if actual != string(expectedBytes) {
		t.Fatalf("Retrieved data not correct.")
	}
}
