package ext4

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestNewBlockGroupDescriptorWithReader(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	bgdOffset := int64(sb.BlockSize() * (sb.Data().SFirstDataBlock + 1))

	_, err = f.Seek(bgdOffset, io.SeekStart)
	log.PanicIf(err)

	bgd, err := NewBlockGroupDescriptorWithReader(f, sb)
	log.PanicIf(err)

	currentPosition, err := f.Seek(0, io.SeekCurrent)
	log.PanicIf(err)

	actualConsumedBytes := currentPosition - bgdOffset

	if actualConsumedBytes != int64(BlockGroupDescriptorSize) {
		t.Fatalf("BGD parse did not consume the right amount of data: (%d) != (%d)", actualConsumedBytes, BlockGroupDescriptorSize)
	}

	if bgd.Data().BgChecksum != 0xeeda {
		t.Fatalf("Checksum not correct.")
	}
}
