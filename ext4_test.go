package ext4

import (
	"os"
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestParse(t *testing.T) {
	filepath := path.Join(assetsPath, "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	err = ParseHead(f)
	log.PanicIf(err)
}
