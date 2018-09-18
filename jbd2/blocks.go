package jbd2

import (
	"fmt"
	"time"
)

type JournalCommonBlockType struct {
	header *JournalHeader
}

func (jcbt *JournalCommonBlockType) SetHeader(jh *JournalHeader) {
	jcbt.header = jh
}

func (jcbt *JournalCommonBlockType) Header() *JournalHeader {
	return jcbt.header
}

const (
	// JBD2_CHECKSUM_BYTES
	Jbd2ChecksumBytes = 8
)

// Descriptor block tag record flags.
const (
	JbtfDataMatchesMagicBytes = uint16(0x1) // On-disk block is escaped. The first four bytes of the data block just happened to match the jbd2 magic number.
	JbtfSameUuidAsPrevious    = uint16(0x2) // This block has the same UUID as previous, therefore the UUID field is omitted.
	JbtfDeleted               = uint16(0x4) // The data block was deleted by the transaction. (Not used?)
	JbtfLastTag               = uint16(0x8) // This is the last tag in this descriptor block.
)

// JournalBlockTagNoCsumV3 (journal_block_tag_s struct) is a subelement of the
// descriptor block.
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

// JournalDescriptorBlock is a top-level journal block-type that describes
// where the data is supposed to go on disk.
type JournalDescriptorBlock struct {
	Tags            []JournalBlockTagNoCsumV3
	transactionData []byte

	JournalCommonBlockType
}

func (jdb *JournalDescriptorBlock) Type() uint32 {
	return BtDescriptor
}

func (jdb *JournalDescriptorBlock) SetTransactionData(transactionData []byte) {
	jdb.transactionData = transactionData
}

func (jdb *JournalDescriptorBlock) String() string {
	return fmt.Sprintf("DescriptorBlock<TAGS=(%d) DATA-LENGTH=(%d)>", len(jdb.Tags), len(jdb.transactionData))
}

func (jdb *JournalDescriptorBlock) Dump() {
	fmt.Printf("%s\n", jdb)
	fmt.Printf("\n")

	for i, jbtnc3 := range jdb.Tags {
		fmt.Printf("  TAG(%d): %s\n", i, jbtnc3)
	}

	fmt.Printf("\n")
}

// JournalCommitBlock (commit_header struct) indicates that a transaction has
// been completely written to the journal.
type JournalCommitBlockData struct {
	// 0xC
	HChksumType uint8 // The type of checksum to use to verify the integrity of the data blocks in the transaction. See jbd2_checksum_type for more info.

	// 0xD
	HChksumSize uint8 // The number of bytes used by the checksum. Most likely 4.

	// 0xE
	HPadding [2]byte

	// 0x10
	HChksum [Jbd2ChecksumBytes]uint32 // 32 bytes of space to store checksums. If JBD2_FEATURE_INCOMPAT_CSUM_V2 or JBD2_FEATURE_INCOMPAT_CSUM_V3 are set, the first __be32 is the checksum of the journal UUID and the entire commit block, with this field zeroed. If JBD2_FEATURE_COMPAT_CHECKSUM is set, the first __be32 is the crc32 of all the blocks already written to the transaction.

	// 0x30
	HCommitSec uint64 // The time that the transaction was committed, in seconds since the epoch.

	// 0x38
	HCommitNsec uint32 // Nanoseconds component of the above timestamp.
}

type JournalCommitBlock struct {
	data *JournalCommitBlockData

	JournalCommonBlockType
}

func (jcb *JournalCommitBlock) Data() *JournalCommitBlockData {
	return jcb.data
}

func (jcb *JournalCommitBlock) Type() uint32 {
	return BtBlockCommitRecord
}

func (jcb *JournalCommitBlock) CommitTime() time.Time {
	return time.Unix(int64(jcb.data.HCommitSec), int64(jcb.data.HCommitNsec))
}

func (jcb *JournalCommitBlock) String() string {
	return fmt.Sprintf("CommitBlock<HChksumType=(%d) HChksumSize=(%d) CommitTime=[%s]>", jcb.data.HChksumType, jcb.data.HChksumSize, jcb.CommitTime())
}
