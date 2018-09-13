package ext4

import (
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestNewWithReadSeeker(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	err = NewExt4WithReadSeeker(f)
	log.PanicIf(err)
}
