package ext4

import (
	"io"

	"github.com/dsoprea/go-logging"
)

// Parse parses the whole filesystem.
func Parse(rs io.ReadSeeker) (err error) {
	_, err = rs.Seek(Superblock0Offset, io.SeekStart)
	log.PanicIf(err)

	sb, esb, err := ParseSuperblock(rs)
	log.PanicIf(err)

	// TODO(dustin): !! Add more. Not very useful, yet.
	sb = sb
	esb = esb

	return nil
}
