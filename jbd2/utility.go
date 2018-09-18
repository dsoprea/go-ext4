package jbd2

import (
	"fmt"
	"io"

	"github.com/dsoprea/go-logging"
)

func DumpBytes(data []byte) {
	fmt.Printf("DUMP: ")
	for _, x := range data {
		fmt.Printf("%02x ", x)
	}

	fmt.Printf("\n")
}

func ReadExactly(r io.Reader, buffer []byte) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	for offset := 0; offset < len(buffer); {
		n, err := r.Read(buffer[offset:])
		log.PanicIf(err)

		offset += n
	}

	return nil
}
