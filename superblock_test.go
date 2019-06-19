package ext4

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"testing"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

func TestNewSuperblockWithReader(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	// Skip over the boot-code at the front of the filesystem.
	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	originalPosition, err := f.Seek(0, io.SeekCurrent)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	currentPosition, err := f.Seek(0, io.SeekCurrent)
	log.PanicIf(err)

	actualConsumedBytes := currentPosition - originalPosition

	if actualConsumedBytes != int64(SuperblockSize) {
		t.Fatalf("Superblock parse did not consume the right amount of data: (%d) != (%d)", actualConsumedBytes, SuperblockSize)
	}

	if sb.Data().SInodesCount != 128 {
		t.Fatalf("SInodesCount not correct.")
	} else if sb.Data().SBlocksCountLo != 1024 {
		t.Fatalf("SBlocksCountLo not correct.")
	} else if sb.Data().SLogBlockSize != 0 {
		t.Fatalf("SLogBlockSize not correct.")
	}

	if sb.Data().SVolumeName != [16]byte{'t', 'i', 'n', 'y', 'i', 'm', 'a', 'g', 'e', 0, 0, 0, 0, 0, 0, 0} {
		t.Fatalf("SVolumeName not correct.")
	} else if sb.Data().SUuid != [16]byte{0x99, 0x05, 0xd1, 0xc3, 0x34, 0x5a, 0x4d, 0xeb, 0x82, 0xb2, 0x49, 0x92, 0xf3, 0xf5, 0x46, 0xcc} {
		t.Fatalf("SUuid not correct.")
	} else if sb.Data().SMkfsTime != 1536385726 {
		t.Fatalf("SMkfsTime not correct.")
	} else if sb.Data().SHashSeed != [4]uint32{0x36b23193, 0x8241e711, 0x56e1ab9, 0x8cb728de} {
		t.Fatalf("SHashSeed not correct.")
	}
}

func ExampleNewSuperblockWithReader() {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	// Skip over the boot-code at the front of the filesystem.
	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	fmt.Println(sb.VolumeName())

	// Output:
	// tinyimage
}

func TestSuperblock_ReadPhysicalBlock(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	pBlock := uint64(sb.Data().SFirstDataBlock)
	data, err := sb.ReadPhysicalBlock(pBlock, uint64(SuperblockSize))
	log.PanicIf(err)

	// Confirm that the read data is the right size.

	if uint64(len(data)) != uint64(SuperblockSize) {
		t.Fatalf("Read data is not the right size: (%d) != (%d)", len(data), SuperblockSize)
	}

	// See if the data parses as a SB and matches the one we already have (as a
	// weak form of validation).

	recoveredSbd := new(SuperblockData)

	b := bytes.NewBuffer(data)

	err = binary.Read(b, binary.LittleEndian, recoveredSbd)
	log.PanicIf(err)

	if reflect.DeepEqual(recoveredSbd, sb.Data()) == false {
		t.Fatalf("Read block was not correct.")
	}
}

func ExampleSuperblock_ReadPhysicalBlock() {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	pBlock := uint64(sb.Data().SFirstDataBlock)
	data, err := sb.ReadPhysicalBlock(pBlock, uint64(SuperblockSize))
	log.PanicIf(err)

	data = data

	// Output:
}

func TestSuperblock_FreeBlockCount(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	_, err = f.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, err := NewSuperblockWithReader(f)
	log.PanicIf(err)

	if sb.FreeBlockCount() != 155 {
		t.Fatalf("Supoerblock free blocks count was incorrect.")
	}
}
