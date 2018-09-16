package ext4

import (
	"io"
	"math"

	"github.com/dsoprea/go-logging"
)

// InodeReader fulfills the `io.Reader` interface to read arbitrary amounts of
// data.
type InodeReader struct {
	en           *ExtentNavigator
	currentBlock []byte
	bytesRead    uint64
	bytesTotal   uint64
}

func NewInodeReader(en *ExtentNavigator) *InodeReader {
	return &InodeReader{
		en:           en,
		currentBlock: make([]byte, 0),
		bytesTotal:   en.inode.Size(),
	}
}

// Read fills the given slice with data and returns an `io.EOF` error with (0)
// bytes when done. (`n`) may be less then `len(p)`.
func (ir *InodeReader) Read(p []byte) (n int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	if len(ir.currentBlock) == 0 {
		if ir.bytesRead >= ir.bytesTotal {
			return 0, io.EOF
		}

		data, err := ir.en.Read(ir.bytesRead)
		log.PanicIf(err)

		ir.currentBlock = data
		ir.bytesRead += uint64(len(data))
	}

	// Determine how much of the buffer we can fill.
	currentBytesReadCount := uint64(math.Min(float64(len(ir.currentBlock)), float64(len(p))))

	copy(p, ir.currentBlock[:currentBytesReadCount])
	ir.currentBlock = ir.currentBlock[currentBytesReadCount:]

	return int(currentBytesReadCount), nil
}
