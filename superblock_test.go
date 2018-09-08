package ext4

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestParseSuperblock(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	//
	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, esb, err := ParseSuperblock(f)
	log.PanicIf(err)

	if sb.SInodesCount != 128 {
		t.Fatalf("SInodesCount not correct.")
	} else if sb.SBlocksCountLo != 1024 {
		t.Fatalf("SBlocksCountLo not correct.")
	} else if sb.SLogBlockSize != 0 {
		t.Fatalf("SLogBlockSize not correct.")
	}

	if esb.SVolumeName != [16]byte{'t', 'i', 'n', 'y', 'i', 'm', 'a', 'g', 'e', 0, 0, 0, 0, 0, 0, 0} {
		t.Fatalf("SVolumeName not correct.")
	} else if esb.SUuid != [16]byte{0x99, 0x05, 0xd1, 0xc3, 0x34, 0x5a, 0x4d, 0xeb, 0x82, 0xb2, 0x49, 0x92, 0xf3, 0xf5, 0x46, 0xcc} {
		t.Fatalf("SUuid not correct.")
	} else if esb.SMkfsTime != 1536385726 {
		t.Fatalf("SMkfsTime not correct.")
	} else if esb.SHashSeed != [4]uint32{0x36b23193, 0x8241e711, 0x56e1ab9, 0x8cb728de} {
		t.Fatalf("SHashSeed not correct.")
	}

	// TODO(dustin): !! Debugging.
	sb = sb
}
