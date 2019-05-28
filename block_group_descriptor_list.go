package ext4

import (
	"io"

	"github.com/dsoprea/go-logging"
)

type BlockGroupDescriptorList struct {
	sb   *Superblock
	bgds []*BlockGroupDescriptor
}

// NewBlockGroupDescriptorListWithReadSeeker returns a
// `BlockGroupDescriptorsList`, which has all block-group-descriptors in a big
// slice. Filesystems with the flex_bg capability flag (most) will group all of
// the BGD data together right at the top.
func NewBlockGroupDescriptorListWithReadSeeker(rs io.ReadSeeker, sb *Superblock) (bgdl *BlockGroupDescriptorList, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// QUESTION(dustin): This whole group is replicated/backed-up along with the superblock?

	// currentBlock initially points at the block with the first BGD.
	initialBlock := uint64(sb.Data().SFirstDataBlock) + 1
	initialOffset := initialBlock * uint64(sb.BlockSize())

	_, err = rs.Seek(int64(initialOffset), io.SeekStart)
	log.PanicIf(err)

	blockGroupsCount := sb.BlockGroupCount()
	bgds := make([]*BlockGroupDescriptor, blockGroupsCount)

	for i := uint64(0); i < blockGroupsCount; i++ {
		bgd, err := NewBlockGroupDescriptorWithReader(rs, sb)
		log.PanicIf(err)

		bgds[i] = bgd
	}

	bgdl = &BlockGroupDescriptorList{
		sb:   sb,
		bgds: bgds,
	}

	return bgdl, nil
}

func (bgdl *BlockGroupDescriptorList) GetWithAbsoluteInode(n int) (bgd *BlockGroupDescriptor, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	blockGroupNumber := bgdl.sb.BlockGroupNumberWithAbsoluteInodeNumber(n)
	return bgdl.bgds[blockGroupNumber], nil
}
