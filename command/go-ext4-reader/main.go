package main

import (
	"os"

	"github.com/dsoprea/go-logging"
	"github.com/jessevdk/go-flags"

	"github.com/dsoprea/go-ext4"
)

var (
	options struct {
		Filepath string `short:"f" long:"filepath" required:"true" description:"EXT4 file/device"`
	}
)

func main() {
	_, err := flags.Parse(&options)
	if err != nil {
		os.Exit(1)
	}

	f, err := os.Open(options.Filepath)
	log.PanicIf(err)

	defer f.Close()

	err = ext4.ParseHead(f)
	log.PanicIf(err)

	// TODO(dustin): !! Add more once we implement more.
}
