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

// JournalHeader (journal_header_s) is at the beginning of every block.
type JournalHeader struct {
	HMagic     uint32
	HBlocktype uint32
	HSequence  uint32
}

func (jh JournalHeader) String() string {
	return fmt.Sprintf("JournalHeader<MAGIC=[%08x] BLOCK-TYPE=[%s] SEQ=(%d)>", jh.HMagic, BlocktypeLookup[jh.HBlocktype], jh.HSequence)
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
	fmt.Printf("SSequence: (%d)\n", jsb.data.SSequence)
	fmt.Printf("SStart: (%d)\n", jsb.data.SStart)
	fmt.Printf("SErrno: (%d)\n", jsb.data.SErrno)
	fmt.Printf("SFeatureCompat: (%d)\n", jsb.data.SFeatureCompat)
	fmt.Printf("SFeatureIncompat: (%d)\n", jsb.data.SFeatureIncompat)
	fmt.Printf("SFeatureRoCompat: (%d)\n", jsb.data.SFeatureRoCompat)
	fmt.Printf("SUuid: (%032x)\n", jsb.data.SUuid)

	fmt.Printf("SNrUsers: (%d)\n", jsb.data.SNrUsers)
	fmt.Printf("SDynsuper: (%d)\n", jsb.data.SDynsuper)
	fmt.Printf("SMaxTransaction: (%d)\n", jsb.data.SMaxTransaction)
	fmt.Printf("SMaxTransData: (%d)\n", jsb.data.SMaxTransData)

	fmt.Printf("SChecksumType: (%d)\n", jsb.data.SChecksumType)
	fmt.Printf("SChecksum: (%d)\n", jsb.data.SChecksum)

	fmt.Printf("\n")
}

type JournalBlock interface {
	Type() uint32
	String() string
	Header() *JournalHeader
}

func (jsb *JournalSuperblock) NextBlock(r io.Reader) (jb JournalBlock, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// Find the next populated block. We've observed sparsity with the populated
	// ones.

	blockSize := int(jsb.data.SBlocksize)

	if jsb.currentBlock >= int(jsb.data.SMaxlen) {
		return nil, io.EOF
	}

	jh := new(JournalHeader)

	err = binary.Read(r, binary.BigEndian, jh)
	log.PanicIf(err)

	remainingBytes := blockSize - JournalHeaderSize
	buffer := make([]byte, remainingBytes)

	err = ReadExactly(r, buffer)
	log.PanicIf(err)

	jsb.currentBlock++

	if jh.HMagic != JournalBlockHeaderMagicBytes {
		// There's no block-type connoting terminating, and the magic-bytes
		// seem to potentially be non-zero even though everything else looks
		// good. Therefore, we're just going to iterate until the magic-bytes
		// are no longer correct.

		// log.Panicf("next block header magic-bytes not correct: %08x", jh.HMagic)
		return nil, io.EOF
	} else if jh.HBlocktype == BtJournalSuperblockV1 || jh.HBlocktype == BtJournalSuperblockV2 {
		log.Panicf("encountered more than one journal superblock")
	}

	remainingReader := bytes.NewBuffer(buffer)

	if jh.HBlocktype == BtDescriptor {
		jdb := new(JournalDescriptorBlock)
		jdb.SetHeader(jh)

		jdb.Tags = make([]JournalBlockTag32NoCsumV3, 0)

		hasLast := false
		i := 0
		for remainingBytes > 0 {
			jbtnc3 := JournalBlockTag32NoCsumV3{}

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TBlocknr)
			log.PanicIf(err)

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TChecksum)
			log.PanicIf(err)

			err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.TFlags)
			log.PanicIf(err)

			remainingBytes -= 8

			if (jbtnc3.TFlags & JbtfSameUuidAsPrevious) == 0 {
				err = binary.Read(remainingReader, binary.BigEndian, &jbtnc3.Uuid)
				log.PanicIf(err)

				remainingBytes -= 4
			}

			jdb.Tags = append(jdb.Tags, jbtnc3)

			// TODO(dustin): !! We're not sure why the commit-tag record has a target block-number on it.
			// TODO(dustin): !! Why don't we see the associated data?
			if (jbtnc3.TFlags & JbtfLastTag) > 0 {
				hasLast = true
				break
			}

			i++
		}

		if hasLast == false {
			log.Panicf("journal descriptor tag blocks not terminated")
		}

		// The next block will have the actual data.

		transactionData := make([]byte, blockSize)

		err = ReadExactly(r, transactionData)
		log.PanicIf(err)

		jdb.SetTransactionData(transactionData)

		return jdb, nil
	} else if jh.HBlocktype == BtBlockCommitRecord {
		jcbd := new(JournalCommitBlockData)

		err := binary.Read(remainingReader, binary.BigEndian, jcbd)
		log.PanicIf(err)

		jcb := &JournalCommitBlock{
			data: jcbd,
		}

		jcb.SetHeader(jh)

		return jcb, nil
	} else if jh.HBlocktype == BtBlockRevocationRecord {
		jrbd := new(JournalRevokeBlock32Data)

		err := binary.Read(remainingReader, binary.BigEndian, jrbd)
		log.PanicIf(err)

		jrb := &JournalRevokeBlock{
			data: jrbd,
		}

		jrb.SetHeader(jh)

		return jrb, nil
	}

	log.Panicf("block-type (%d) not handled", jh.HBlocktype)
	return nil, nil
}
