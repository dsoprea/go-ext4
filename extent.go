package ext4

import (
	"bytes"
	"fmt"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	ExtentMagic            = uint16(0xf30A)
	ExtentHeaderSize       = 12
	ExtentIndexAndLeafSize = 12
)

type ExtentHeaderNode struct {
	EhMagic      uint16 /* probably will support different formats */
	EhEntryCount uint16 /* number of valid entries */
	EhMax        uint16 /* capacity of store in entries */
	EhDepth      uint16 /* has tree real underlying blocks? */
	EhGeneration uint32 /* generation of the tree */
}

func (eh *ExtentHeaderNode) String() string {
	return fmt.Sprintf("ExtentHeaderNode<ENTRIES=(%d) MAX=(%d) DEPTH=(%d)>", eh.EhEntryCount, eh.EhMax, eh.EhDepth)
}

type ExtentIndexNode struct {
	EiLogicalBlock        uint32 /* index covers logical blocks from 'block' */
	EiLeafPhysicalBlockLo uint32 /* pointer to the physical block of the next level. leaf or next index could be there */
	EiLeafPhysicalBlockHi uint16 /* high 16 bits of physical block */
	EiUnused              uint16
}

func (ein *ExtentIndexNode) LeafPhysicalBlock() uint64 {
	return uint64(ein.EiLeafPhysicalBlockHi<<32) | uint64(ein.EiLeafPhysicalBlockLo)
}

func (ein *ExtentIndexNode) String() string {
	return fmt.Sprintf("ExtentIndexNode<FILE-LBLOCK=(%d) LEAF-PBLOCK=(%d)>", ein.EiLogicalBlock, ein.LeafPhysicalBlock())
}

type ExtentLeafNode struct {
	EeFirstLogicalBlock    uint32 /* first logical block extent covers */
	EeLogicalBlockCount    uint16 /* number of blocks covered by extent */
	EeStartPhysicalBlockHi uint16 /* high 16 bits of physical block */
	EeStartPhysicalBlockLo uint32 /* low 32 bits of physical block */
}

func (eln *ExtentLeafNode) StartPhysicalBlock() uint64 {
	return uint64(eln.EeStartPhysicalBlockHi<<32) | uint64(eln.EeStartPhysicalBlockLo)
}

func (eln *ExtentLeafNode) String() string {
	return fmt.Sprintf("ExtentLeafNode<FIRST-LBLOCK=(%d) LBLOCK-COUNT=(%d) START-PBLOCK=(%d)>", eln.EeFirstLogicalBlock, eln.EeLogicalBlockCount, eln.StartPhysicalBlock())
}

type ExtentTail struct {
	EbChecksum uint32
}

type ExtentNavigator struct {
	rs    io.ReadSeeker
	inode *Inode
}

func NewExtentNavigatorWithReadSeeker(rs io.ReadSeeker, inode *Inode) *ExtentNavigator {
	return &ExtentNavigator{
		rs:    rs,
		inode: inode,
	}
}

// Block returns a physical-block number for a specific logical-block of data
// for the principal inode.
//
// "logical", meaning that (0) refers to the first block of this inode's data.
func (en *ExtentNavigator) Read(offset uint64) (data []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	sb := en.inode.BlockGroupDescriptor().Superblock()

	blockSize := uint64(sb.BlockSize())
	lBlockNumber := offset / blockSize
	pBlockOffset := offset % blockSize

	inodeIblock := en.inode.Data().IBlock[:]
	pBlockNumber, err := en.parseHeader(inodeIblock, lBlockNumber)
	log.PanicIf(err)

	// We'll return whichever data we got between the offset and the end of
	// that immediate physical block.
	rawPBlockData, err := sb.ReadPhysicalBlock(pBlockNumber, blockSize)
	log.PanicIf(err)

	return rawPBlockData[pBlockOffset:], nil
}

// parseHeader parses the extent header and then recursively processes the
// array of index-nodes or array of leaf-nodes following it.
func (en *ExtentNavigator) parseHeader(extentHeaderData []byte, lBlock uint64) (dataPBlock uint64, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	b := bytes.NewBuffer(extentHeaderData)

	// TODO(dustin): Pass this in as another argument and only parse if we receive a nil. Except for the first one, we'll otherwise double-parse every header struct.
	eh := new(ExtentHeaderNode)

	err = binary.Read(b, binary.LittleEndian, eh)
	log.PanicIf(err)

	if eh.EhMagic != ExtentMagic {
		log.Panicf("extent-header magic-bytes not correct: (%04x)", eh.EhMagic)
	}

	if eh.EhDepth == 0 {
		// Our nodes are leaf nodes.

		leafNodes := make([]ExtentLeafNode, eh.EhEntryCount)

		err = binary.Read(b, binary.LittleEndian, &leafNodes)
		log.PanicIf(err)

		// Forward through the leaf-nodes on this level until we find one that
		// extends beyond the logical-block we wanted.

		var hit *ExtentLeafNode
		for i, eln := range leafNodes {
			if uint64(eln.EeFirstLogicalBlock+uint32(eln.EeLogicalBlockCount)) > lBlock {
				hit = &leafNodes[i]
				break
			}
		}

		blockExtOffset := lBlock - uint64(hit.EeFirstLogicalBlock)
		pBlock := hit.StartPhysicalBlock() + blockExtOffset

		return pBlock, nil
	} else {
		// Our nodes are interior/index nodes.

		indexNodes := make([]ExtentIndexNode, eh.EhEntryCount)

		err = binary.Read(b, binary.LittleEndian, &indexNodes)
		log.PanicIf(err)

		var hit *ExtentIndexNode
		for i, ein := range indexNodes {
			if uint64(ein.EiLogicalBlock) <= lBlock {
				hit = &indexNodes[i]
			} else {
				break
			}
		}

		if hit == nil {
			log.Panicf("None of the index nodes at the current level of the "+
				"extent-tree for inode (%d) had a logical-block less "+
				"than what was requested (%d).", en.inode, lBlock)
		}

		pBlock := hit.LeafPhysicalBlock()

		// TODO(dustin): Refactor this to prevent reparsing the data in the next recursion when we're already parsing it here.

		// Do a preliminary read of the header to establish how much data we
		// really need.

		sb := en.inode.BlockGroupDescriptor().Superblock()

		data, err := sb.ReadPhysicalBlock(pBlock, uint64(ExtentHeaderSize))
		log.PanicIf(err)

		nonleafHeaderBuffer := bytes.NewBuffer(data)

		nextEh := new(ExtentHeaderNode)

		err = binary.Read(nonleafHeaderBuffer, binary.LittleEndian, nextEh)
		log.PanicIf(err)

		// Now, read the full data for our child extents.

		childExtentsLength := ExtentHeaderSize + ExtentIndexAndLeafSize*nextEh.EhEntryCount

		childExtents, err := sb.ReadPhysicalBlock(pBlock, uint64(childExtentsLength))
		log.PanicIf(err)

		dataPBlock, err = en.parseHeader(childExtents, lBlock)
		log.PanicIf(err)

		return dataPBlock, nil
	}
}
