package ext4

// TODO(dustin): !! We don't yet see a lot of value for this. We might remove it in the future.

import (
	"io"

	"github.com/dsoprea/go-logging"
)

type Ext4Parser struct {
	sb  *Superblock
	bgd *BlockGroupDescriptor

	blockSize uint32
}

func (ep *Ext4Parser) Superblock() *Superblock {
	return ep.sb
}

func (ep *Ext4Parser) BlockGroupDescriptor() *BlockGroupDescriptor {
	return ep.bgd
}

func (ep *Ext4Parser) BlockOffset(n uint64) uint64 {
	if ep.blockSize == 1024 {
		n++
	}

	return uint64(ep.blockSize) * n
}

func (ep *Ext4Parser) SeekToBlock(rs io.ReadSeeker, n uint64) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	offset := ep.BlockOffset(n)

	_, err = rs.Seek(int64(offset), io.SeekStart)
	log.PanicIf(err)

	return nil
}

func NewExt4ParserFromReadSeeker(rs io.ReadSeeker, isFirst bool) (ep *Ext4Parser, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// If we're reading the first superblock and BGD in the filesystem, we'll
	// have to seek back the bootcode.
	if isFirst == true {
		_, err := rs.Seek(Superblock0Offset, io.SeekStart)
		log.PanicIf(err)
	}

	sb, err := NewSuperblockWithReader(rs)
	log.PanicIf(err)

	blockSize := sb.BlockSize()

	ep = &Ext4Parser{
		sb:        sb,
		blockSize: blockSize,
	}

	// If we're still in the middle of the block that hosts the superblock,
	// jump to the next.
	if blockSize > SuperblockSize {
		_, err := rs.Seek(int64(blockSize-SuperblockSize), io.SeekCurrent)
		log.PanicIf(err)
	}

	bgd, err := NewBlockGroupDescriptorWithReader(rs, sb)
	log.PanicIf(err)

	ep.bgd = bgd

	return ep, nil
}
