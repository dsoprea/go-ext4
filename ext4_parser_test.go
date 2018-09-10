package ext4

import (
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestNewExt4ParserFromReadSeeker(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	ep, err := NewExt4ParserFromReadSeeker(f, true)
	log.PanicIf(err)

	ep.Superblock()
	ep.BlockGroupDescriptor()
}
