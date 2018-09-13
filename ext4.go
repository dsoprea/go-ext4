package ext4

import (
	"io"

	"github.com/dsoprea/go-logging"
)

// NewExt4WithReadSeeker parses the first superblock and the first block group
// descriptor. Everything after that needs needs seeking and needs integrity.
func NewExt4WithReadSeeker(rs io.ReadSeeker) (err error) {
	ep, err := NewExt4ParserFromReadSeeker(rs, true)
	log.PanicIf(err)

	// TODO(dustin): !! Add more. Not very useful, yet.

	ep = ep
	// ep.Superblock().Dump()
	// ep.BlockGroupDescriptor().Dump()

	return nil
}
