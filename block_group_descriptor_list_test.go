package ext4

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestNewBlockGroupDescriptorListWithReadSeeker(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	// Skip over the boot-code at the front of the filesystem.
	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdl, err := NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	log.PanicIf(err)

	if len(bgdl.bgds) != 1 {
		t.Fatalf("Expected exactly one BGD: (%d)", len(bgdl.bgds))
	} else if bgdl.bgds[0].Data().BgChecksum != 0xeeda {
		t.Fatalf("BGD checksum is not correct: [%04x]", bgdl.bgds[0].Data().BgChecksum)
	}
}
