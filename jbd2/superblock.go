package jbd2

import (
	"bytes"
	"fmt"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	JournalBlockHeaderMagicBytes = uint32(0xc03b3998)
	JournalHeaderSize            = 12
)

// Block types
const (
	BtDescriptor            = uint32(1) // Descriptor. This block precedes a series of data blocks that were written through the journal during a transaction.
	BtBlockCommitRecord     = uint32(2) // Block commit record. This block signifies the completion of a transaction.
	BtJournalSuperblockV1   = uint32(3) // Journal superblock, v1.
	BtJournalSuperblockV2   = uint32(4) // Journal superblock, v2.
	BtBlockRevocationRecord = uint32(5) // Block revocation records. This speeds up recovery by enabling the journal to skip writing blocks that were subsequently rewritten.
)

var (
	BlocktypeLookup = map[uint32]string{
		BtDescriptor:            "descriptor",
		BtBlockCommitRecord:     "block commit record",
		BtJournalSuperblockV1:   "journal superblock v1",
		BtJournalSuperblockV2:   "journal superblock v2",
		BtBlockRevocationRecord: "block revocation record",
	}
)

// Journal superblock "compat" features.
const (
	// JBD2_FEATURE_COMPAT_CHECKSUM
	JsbFeatureCompatChecksum = uint32(0x1)
)

var (
	JsbFeatureCompatLookup = map[uint32]string{
		JsbFeatureCompatChecksum: "checksum",
	}
)

// Journal superblock "incomp" features.
const (
	// JBD2_FEATURE_INCOMPAT_REVOKE
	JsbFeatureIncompatRevoke = uint32(0x1) // Journal has block revocation records.

	// JBD2_FEATURE_INCOMPAT_64BIT
	JsbFeatureIncompat64bit = uint32(0x2) // Journal can deal with 64-bit block numbers.

	// JBD2_FEATURE_INCOMPAT_ASYNC_COMMIT
	JsbFeatureIncompatAsyncCommit = uint32(0x4) // Journal commits asynchronously.

	// JBD2_FEATURE_INCOMPAT_CSUM_V2
	JsbFeatureIncompatCsumV2 = uint32(0x8) // This journal uses v2 of the checksum on-disk format. Each journal metadata block gets its own checksum, and the block tags in the descriptor table contain checksums for each of the data blocks in the journal.

	// JBD2_FEATURE_INCOMPAT_CSUM_V3
	JsbFeatureIncompatCsumV3 = uint32(0x10) // This journal uses v3 of the checksum on-disk format. This is the same as v2, but the journal block tag size is fixed regardless of the size of block numbers.
)

var (
	JsbFeatureIncompatLookup = map[uint32]string{
		JsbFeatureIncompatRevoke:      "revoke",
		JsbFeatureIncompat64bit:       "64bit",
		JsbFeatureIncompatAsyncCommit: "asynccommit",
		JsbFeatureIncompatCsumV2:      "csumv2",
		JsbFeatureIncompatCsumV3:      "csumv3",
	}
)

// Journal checksum codes
const (
	JccCrc32  = uint8(1)
	JccMd5    = uint8(2)
	JccSha1   = uint8(3)
	JccCrc32c = uint8(4)
)

type JournalHeader struct {
	HMagic     uint32
	HBlocktype uint32
	HSequence  uint32
}

func (jh JournalHeader) String() string {
	return fmt.Sprintf("JournalHeader<BLOCK-TYPE=[%s]>", BlocktypeLookup[jh.HBlocktype])
}

func (jh *JournalHeader) Dump() {
	fmt.Printf("Journal Superblock\n")
	fmt.Printf("==================\n")
	fmt.Printf("\n")

	fmt.Printf("HMagic: %08x\n", jh.HMagic)
	fmt.Printf("HBlocktype: %d\n", jh.HBlocktype)
	fmt.Printf("HSequence: %d\n", jh.HSequence)

	fmt.Printf("\n")
}

type JournalSuperblockData struct {
	/* 0x0000 */
	SHeader JournalHeader

	/* 0x000C */
	/* Static information describing the journal */
	SBlocksize uint32 /* journal device blocksize */
	SMaxlen    uint32 /* total blocks in journal file */
	SFirst     uint32 /* first block of log information */

	/* 0x0018 */
	/* Dynamic information describing the current state of the log */
	SSequence uint32 /* first commit ID expected in log */
	SStart    uint32 /* blocknr of start of log */

	/* 0x0020 */
	/* Error value, as set by jbd2_journal_abort(). */
	SErrno uint32

	/* 0x0024 */
	/* Remaining fields are only valid in a version-2 superblock */
	SFeatureCompat   uint32 /* compatible feature set */
	SFeatureIncompat uint32 /* incompatible feature set */
	SFeatureRoCompat uint32 /* readonly-compatible feature set */

	/* 0x0030 */
	SUuid [16]uint8 /* 128-bit uuid for journal */

	/* 0x0040 */
	SNrUsers uint32 /* Nr of filesystems sharing log */

	SDynsuper uint32 /* Blocknr of dynamic superblock copy*/

	/* 0x0048 */
	SMaxTransaction uint32 /* Limit of journal blocks per trans.*/
	SMaxTransData   uint32 /* Limit of data blocks per trans. */

	/* 0x0050 */
	SChecksumType uint8 /* checksum type */
	SPadding2     [3]uint8
	SPadding      [42]uint32
	SChecksum     uint32 /* crc32c(superblock) */

	/* 0x0100 */
	SUsers [16 * 48]uint8 /* ids of all fs'es sharing the log */

	/* 0x0400 */
}

type JournalSuperblock struct {
	data         *JournalSuperblockData
	currentBlock int
}

func NewJournalSuperblock(r io.Reader) (jsb *JournalSuperblock, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	jsbd := new(JournalSuperblockData)

	err = binary.Read(r, binary.BigEndian, jsbd)
	log.PanicIf(err)

	if jsbd.SHeader.HMagic != JournalBlockHeaderMagicBytes {
		log.Panicf("journal header magic-bytes not correct: %08x", jsbd.SHeader.HMagic)
	} else if jsbd.SHeader.HBlocktype != BtJournalSuperblockV2 {
		log.Panicf("only superblock V2 supported")
	}

	jsb = &JournalSuperblock{
		data: jsbd,
	}

	if jsb.HasIncompatibleFeature(JsbFeatureIncompatRevoke) == true {
		log.Panicf("revoke feature not supported")
	} else if jsb.HasIncompatibleFeature(JsbFeatureIncompat64bit) == true {
		log.Panicf("64-bit feature not supported")
	} else if jsb.HasIncompatibleFeature(JsbFeatureIncompatAsyncCommit) == true {
		log.Panicf("async-commit feature not supported")
	} else if jsb.HasIncompatibleFeature(JsbFeatureIncompatCsumV2) == true {
		log.Panicf("csum V2 feature not supported")
	} else if jsb.HasIncompatibleFeature(JsbFeatureIncompatCsumV3) == true {
		log.Panicf("csum V3 feature not supported")
	}

	return jsb, nil
}

func (jsb *JournalSuperblock) Data() *JournalSuperblockData {
	return jsb.data
}

func (jsb *JournalSuperblock) HasCompatibleFeature(mask uint32) bool {
	return (jsb.data.SFeatureCompat & mask) > 0
}

func (jsb *JournalSuperblock) HasIncompatibleFeature(mask uint32) bool {
	return (jsb.data.SFeatureIncompat & mask) > 0
}

func (jsb *JournalSuperblock) DumpFeatures(includeFalses bool) {
	fmt.Printf("Features (Compatible)\n")
	fmt.Printf("\n")

	for bit, name := range JsbFeatureCompatLookup {
		value := jsb.HasCompatibleFeature(bit)

		if includeFalses == true || value == true {
			fmt.Printf("  %15s (0x%02x): %v\n", name, bit, value)
		}
	}

	fmt.Printf("\n")

	fmt.Printf("Features (Incompatible)\n")
	fmt.Printf("\n")

	for bit, name := range JsbFeatureIncompatLookup {
		value := jsb.HasIncompatibleFeature(bit)

		if includeFalses == true || value == true {
			fmt.Printf("  %15s (0x%02x): %v\n", name, bit, value)
		}
	}

	fmt.Printf("\n")
}

func (jsb *JournalSuperblock) Dump() {
	fmt.Printf("Journal Superblock\n")
	fmt.Printf("==================\n")
	fmt.Printf("\n")

	fmt.Printf("SBlocksize: (%d)\n", jsb.data.SBlocksize)
	fmt.Printf("SMaxLen: (%d)\n", jsb.data.SMaxlen)
	fmt.Printf("SFirst: (%d)\n", jsb.data.SFirst)

	// TODO(dustin): !! Finish printing remaining data.

	fmt.Printf("\n")
}

type JournalBlock interface {
	Type() uint32
}

const (
	JbtfDataMatchesMagicBytes = uint16(0x1) // On-disk block is escaped. The first four bytes of the data block just happened to match the jbd2 magic number.
	JbtfSameUuidAsPrevious    = uint16(0x2) // This block has the same UUID as previous, therefore the UUID field is omitted.
	JbtfDeleted               = uint16(0x4) // The data block was deleted by the transaction. (Not used?)
	JbtfLastTag               = uint16(0x8) // This is the last tag in this descriptor block.
)

type JournalBlockTagNoCsumV3 struct {
	// 0x0
	TBlocknr uint32 // Lower 32-bits of the location of where the corresponding data block should end up on disk.

	// 0x4
	TChecksum uint16 // Checksum of the journal UUID, the sequence number, and the data block. Note that only the lower 16 bits are stored.

	// 0x6
	TFlags uint16 // Flags that go with the descriptor. See the table jbd2_tag_flags for more info.

	// Only present if the super block indicates support for 64-bit block numbers.
	// TBlocknrHigh uint32 //  Upper 32-bits of the location of where the corresponding data block should end up on disk.

	// 0x8 or 0xC
	// This field appears to be open coded. It always comes at the end of the tag, after t_flags or t_blocknr_high. This field is not present if the “same UUID” flag is set.
	Uuid [16]byte //  A UUID to go with this tag. This field appears to be copied from the j_uuid field in struct journal_s, but only tune2fs touches that field.
}

func (jbtnc3 JournalBlockTagNoCsumV3) String() string {
	return fmt.Sprintf("JournalBlockTag<TBLOCKNR=(%d) TCHECKSUM=(%d) TFLAGS=(%d) UUID=[%032x]>", jbtnc3.TBlocknr, jbtnc3.TChecksum, jbtnc3.TFlags, jbtnc3.Uuid)
}

type JournalDescriptorBlock struct {
	Tags []JournalBlockTagNoCsumV3
}

func (jdb *JournalDescriptorBlock) Type() uint32 {
	return BtDescriptor
}

func (jsb *JournalSuperblock) NextBlock(r io.Reader) (jb JournalBlock, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	if jsb.currentBlock >= int(jsb.data.SMaxlen) {
		return nil, io.EOF
	}

	jh := new(JournalHeader)

	err = binary.Read(r, binary.BigEndian, jh)
	log.PanicIf(err)

	if jh.HMagic == 0 {
		return nil, io.EOF
	} else if jh.HMagic != JournalBlockHeaderMagicBytes {
		log.Panicf("next block header magic-bytes not correct: %08x", jh.HMagic)
	} else if jh.HBlocktype == BtJournalSuperblockV1 || jh.HBlocktype == BtJournalSuperblockV2 {
		log.Panicf("encountered more than one journal superblock")
	}

	remainingBytes := int(jsb.data.SBlocksize) - JournalHeaderSize
	needBytes := remainingBytes
	buffer := make([]byte, needBytes)
	for offset := 0; needBytes > 0; {
		n, err := r.Read(buffer[offset:])
		log.PanicIf(err)

		offset += n
		needBytes -= n
	}

	remainingReader := bytes.NewBuffer(buffer)

	if jh.HBlocktype == BtDescriptor {
		jdb := new(JournalDescriptorBlock)
		jdb.Tags = make([]JournalBlockTagNoCsumV3, 0)

		hasLast := false
		for remainingBytes > 0 {
			jbtnc3 := JournalBlockTagNoCsumV3{}

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TBlocknr)
			log.PanicIf(err)

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TChecksum)
			log.PanicIf(err)

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TFlags)
			log.PanicIf(err)

			remainingBytes -= 8

			if (jbtnc3.TFlags & JbtfSameUuidAsPrevious) > 0 {
				err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.Uuid)
				log.PanicIf(err)

				remainingBytes -= 4
			}

			jdb.Tags = append(jdb.Tags, jbtnc3)

			if (jbtnc3.TFlags & JbtfLastTag) > 0 {
				hasLast = true
				break
			}
		}

		if hasLast == false {
			log.Panicf("journal descriptor tag blocks not terminated")
		}

		return jdb, nil
	} else if jh.HBlocktype == BtBlockCommitRecord {
		// TODO(dustin): !! Finish.
	} else if jh.HBlocktype == BtBlockRevocationRecord {
		// TODO(dustin): !! Finish.
	}

	log.Panicf("block-type (%d) not handled", jh.HBlocktype)
	return nil, nil
}
