package ext4

import (
	"bytes"
	"fmt"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	ExtentMagic = uint16(0xf30A)
)

type ExtentHeader struct {
	EhMagic      uint16 /* probably will support different formats */
	EhEntries    uint16 /* number of valid entries */
	EhMax        uint16 /* capacity of store in entries */
	EhDepth      uint16 /* has tree real underlying blocks? */
	EhGeneration uint32 /* generation of the tree */
}

func (eh *ExtentHeader) String() string {
	return fmt.Sprintf("ExtentHeader<ENTRIES=(%d) MAX=(%d) DEPTH=(%d)>", eh.EhEntries, eh.EhMax, eh.EhDepth)
}

type ExtentIndexNode struct {
	EiBlock  uint32 /* index covers logical blocks from 'block' */
	EiLeafLo uint32 /* pointer to the physical block of the next level. leaf or next index could be there */
	EiLeafHi uint16 /* high 16 bits of physical block */
	EiUnused uint16
}

type ExtentLeafNode struct {
	EeBlock   uint32 /* first logical block extent covers */
	EeLen     uint16 /* number of blocks covered by extent */
	EeStartHi uint16 /* high 16 bits of physical block */
	EeStartLo uint32 /* low 32 bits of physical block */
}

type ExtentTail struct {
	EbChecksum uint32
}

type ExtentNavigator struct {
	rs    io.ReadSeeker
	inode *Inode
}

func NewExtentNavigator(rs io.ReadSeeker, inode *Inode) *ExtentNavigator {
	return &ExtentNavigator{
		rs:    rs,
		inode: inode,
	}
}

func (en *ExtentNavigator) Block(lblock uint64) (data []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	b := bytes.NewBuffer(en.inode.Data().IBlock[:])

	eh := new(ExtentHeader)

	err = binary.Read(b, binary.LittleEndian, eh)
	log.PanicIf(err)

	if eh.EhMagic != ExtentMagic {
		log.Panicf("extent magic-bytes not correct: (%04x)", eh.EhMagic)
	}

	// TODO(dustin): !! Finish.
	return nil, nil
}
