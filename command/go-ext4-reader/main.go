package main

import (
	"os"
	"path"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-ext4"
)

func main() {
	// TODO(dustin): !! Debugging.
	filepath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "dsoprea", "go-ext4", "assets", "tiny.ext4")

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	err = ext4.Parse(f)
	log.PanicIf(err)

	// TODO(dustin): !! Add more once we implement more.
}
