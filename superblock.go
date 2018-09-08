package ext4

import (
	"errors"
	"fmt"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	Ext4Magic = 0xef53
)

var (
	ErrNotExt4 = errors.New("not ext4")
)

const (
	SbStateCleanlyUnmounted      = 0x1
	SbStateErrorsDetected        = 0x2
	SbStateOrphansBeingRecovered = 0x4
)

const (
	SbErrorsContinue        = 0x1
	SbErrorsRemountReadonly = 0x2
	SbErrorsPanic           = 0x3
)

const (
	SbOsLinux   = 0x0
	SbOsHurd    = 0x1
	SbOsMasix   = 0x2
	SbOsFreebsd = 0x3
	SbOsLites   = 0x4
)

const (
	SbRevlevelGoodOldRev = 0x0
	SbRevlevelDynamicRev = 0x1
)

type Superblock struct {
	// See fs/ext4/ext4.h .

	SInodesCount       uint32
	SBlocksCountLo     uint32
	SRBlocksCountLo    uint32
	SFreeBlocksCountLo uint32
	SFreeInodesCount   uint32
	SFirstDataBlock    uint32
	SLogBlockSize      uint32
	SLogClusterSize    uint32
	SBlocksPerGroup    uint32
	SClustersPerGroup  uint32
	SInodesPerGroup    uint32
	SMtime             uint32
	SWtime             uint32
	SMntCount          uint16
	SMaxMntCount       uint16
	SMagic             uint16
	SState             uint16
	SErrors            uint16
	SMinorRevLevel     uint16
	SLastcheck         uint32
	SCheckinterval     uint32
	SCreatorOs         uint32
	SRevLevel          uint32
	SDefResuid         uint16
	SDefResgid         uint16
}

const (
	SbeFeatureCompatDirPrealloc  = uint32(0x0001)
	SbeFeatureCompatImagicInodes = uint32(0x0002)
	SbeFeatureCompatHasJournal   = uint32(0x0004)
	SbeFeatureCompatExtAttr      = uint32(0x0008)
	SbeFeatureCompatResizeInode  = uint32(0x0010)
	SbeFeatureCompatDirIndex     = uint32(0x0020)
)

var (
	SbeFeatureCompatNames = []string{
		"DirIndex",
		"DirPrealloc",
		"ExtAttr",
		"HasJournal",
		"ImagicInodes",
		"ResizeInode",
	}

	SbeFeatureCompatLookup = map[string]uint32{
		"DirPrealloc":  SbeFeatureCompatDirPrealloc,
		"ImagicInodes": SbeFeatureCompatImagicInodes,
		"HasJournal":   SbeFeatureCompatHasJournal,
		"ExtAttr":      SbeFeatureCompatExtAttr,
		"ResizeInode":  SbeFeatureCompatResizeInode,
		"DirIndex":     SbeFeatureCompatDirIndex,
	}
)

const (
	SbeFeatureRoCompatSparseSuper = uint32(0x0001)
	SbeFeatureRoCompatLargeFile   = uint32(0x0002)
	SbeFeatureRoCompatBtreeDir    = uint32(0x0004)
	SbeFeatureRoCompatHugeFile    = uint32(0x0008)
	SbeFeatureRoCompatGdtCsum     = uint32(0x0010)
	SbeFeatureRoCompatDirNlink    = uint32(0x0020)
	SbeFeatureRoCompatExtraIsize  = uint32(0x0040)
)

var (
	SbeFeatureRoCompatNames = []string{
		"BtreeDir",
		"DirNlink",
		"ExtraIsize",
		"GdtCsum",
		"HugeFile",
		"LargeFile",
		"SparseSuper",
	}

	SbeFeatureRoCompatLookup = map[string]uint32{
		"SparseSuper": SbeFeatureRoCompatSparseSuper,
		"LargeFile":   SbeFeatureRoCompatLargeFile,
		"BtreeDir":    SbeFeatureRoCompatBtreeDir,
		"HugeFile":    SbeFeatureRoCompatHugeFile,
		"GdtCsum":     SbeFeatureRoCompatGdtCsum,
		"DirNlink":    SbeFeatureRoCompatDirNlink,
		"ExtraIsize":  SbeFeatureRoCompatExtraIsize,
	}
)

const (
	SbeFeatureIncompatCompression = uint32(0x0001)
	SbeFeatureIncompatFiletype    = uint32(0x0002)
	SbeFeatureIncompatRecover     = uint32(0x0004) /* Needs recovery */
	SbeFeatureIncompatJournalDev  = uint32(0x0008) /* Journal device */
	SbeFeatureIncompatMetaBg      = uint32(0x0010)
	SbeFeatureIncompatExtents     = uint32(0x0040) /* extents support */
	SbeFeatureIncompat64bit       = uint32(0x0080)
	SbeFeatureIncompatMmp         = uint32(0x0100)
	SbeFeatureIncompatFlexBg      = uint32(0x0200)
)

var (
	SbeFeatureIncompatNames = []string{
		"64bit",
		"Compression",
		"Extents",
		"Filetype",
		"FlexBg",
		"JournalDev",
		"MetaBg",
		"Mmp",
		"Recover",
	}

	SbeFeatureIncompatLookup = map[string]uint32{
		"Compression": SbeFeatureIncompatCompression,
		"Filetype":    SbeFeatureIncompatFiletype,
		"Recover":     SbeFeatureIncompatRecover,
		"JournalDev":  SbeFeatureIncompatJournalDev,
		"MetaBg":      SbeFeatureIncompatMetaBg,
		"Extents":     SbeFeatureIncompatExtents,
		"64bit":       SbeFeatureIncompat64bit,
		"Mmp":         SbeFeatureIncompatMmp,
		"FlexBg":      SbeFeatureIncompatFlexBg,
	}
)

// SuperblockExtension available if `SRevLevel` == `SB_REVLEVEL_V2_DYNAMIC_INODES`.
type SuperblockExtension struct {
	/*
	 * These fields are for EXT4_DYNAMIC_REV superblocks only.
	 *
	 * Note: the difference between the compatible feature set and
	 * the incompatible feature set is that if there is a bit set
	 * in the incompatible feature set that the kernel doesn't
	 * know about, it should refuse to mount the filesystem.
	 *
	 * e2fsck's requirements are more strict; if it doesn't know
	 * about a feature in either the compatible or incompatible
	 * feature set, it must abort and not try to meddle with
	 * things it doesn't understand...
	 */
	SFirstIno             uint32    /* First non-reserved inode */
	SInodeSize            uint16    /* size of inode structure */
	SBlockGroupNr         uint16    /* block group # of this superblock */
	SFeatureCompat        uint32    /* compatible feature set */
	SFeatureIncompat      uint32    /* incompatible feature set */
	SFeatureRoCompat      uint32    /* readonly-compatible feature set */
	SUuid                 [16]uint8 /* 128-bit uuid for volume */
	SVolumeName           [16]byte  /* volume name */
	SLastMounted          [64]byte  /* directory where last mounted */
	SAlgorithmUsageBitmap uint32    /* For compression */
	/*
	 * Performance hints.  Directory preallocation should only
	 * happen if the EXT4_FEATURE_COMPAT_DIR_PREALLOC flag is on.
	 */
	SPreallocBlocks    uint8  /* Nr of blocks to try to preallocate*/
	SPreallocDirBlocks uint8  /* Nr to preallocate for dirs */
	SReservedGdtBlocks uint16 /* Per group desc for online growth */
	/*
	 * Journaling support valid if EXT4_FEATURE_COMPAT_HAS_JOURNAL set.
	 */
	SJournalUuid      [16]uint8 /* uuid of journal superblock */
	SJournalInum      uint32    /* inode number of journal file */
	SJournalDev       uint32    /* device number of journal file */
	SLastOrphan       uint32    /* start of list of inodes to delete */
	SHashSeed         [4]uint32 /* HTREE hash seed */
	SDefHashVersion   uint8     /* Default hash version to use */
	SReservedCharPad  uint8
	SDescSize         uint16 /* size of group descriptor */
	SDefaultMountOpts uint32
	SFirstMetaBg      uint32     /* First metablock block group */
	SMkfsTime         uint32     /* When the filesystem was created */
	SJnlBlocks        [17]uint32 /* Backup of the journal inode */

	/* 64bit support valid if EXT4_FEATURE_COMPAT_64BIT */
	SBlocksCountHi     uint32 /* Blocks count */
	SRBlocksCountHi    uint32 /* Reserved blocks count */
	SFreeBlocksCountHi uint32 /* Free blocks count */
	SMinExtraIsize     uint16 /* All inodes have at least # bytes */
	SWantExtraIsize    uint16 /* New inodes should reserve # bytes */

	SFlags            uint32 /* Miscellaneous flags */
	SRaidStride       uint16 /* RAID stride */
	SMmpInterval      uint16 /* # seconds to wait in MMP checking */
	SMmpBlock         uint16 /* Block for multi-mount protection */
	SRaidStripeWidth  uint32 /* blocks on all data disks (N*stride)*/
	SLogGroupsPerFlex uint8  /* FLEX_BG group size */
	SReservedCharPad2 uint8
	SReservedPad      uint16
	SKbytesWritten    uint64      /* nr of lifetime kilobytes written */
	SReserved         [160]uint32 /* Padding to the end of the block */
}

func (sbe *SuperblockExtension) HasCompatibleFeature(mask uint32) bool {
	return (sbe.SFeatureCompat & mask) > 0
}

func (sbe *SuperblockExtension) HasReadonlyCompatibleFeature(mask uint32) bool {
	return (sbe.SFeatureRoCompat & mask) > 0
}

func (sbe *SuperblockExtension) HasIncompatibleFeature(mask uint32) bool {
	return (sbe.SFeatureIncompat & mask) > 0
}

func (sbe *SuperblockExtension) Dump() {
	fmt.Printf("Feature (Compatible)\n")
	fmt.Printf("\n")

	for _, name := range SbeFeatureCompatNames {
		bit := SbeFeatureCompatLookup[name]
		fmt.Printf("  %15s (0x%02x): %v\n", name, bit, sbe.HasCompatibleFeature(bit))
	}

	fmt.Printf("\n")

	fmt.Printf("Feature (Read-Only Compatible)\n")
	fmt.Printf("\n")

	for _, name := range SbeFeatureRoCompatNames {
		bit := SbeFeatureRoCompatLookup[name]
		fmt.Printf("  %15s (0x%02x): %v\n", name, bit, sbe.HasReadonlyCompatibleFeature(bit))
	}

	fmt.Printf("\n")

	fmt.Printf("Feature (Incompatible)\n")
	fmt.Printf("\n")

	for _, name := range SbeFeatureIncompatNames {
		bit := SbeFeatureIncompatLookup[name]
		fmt.Printf("  %15s (0x%02x): %v\n", name, bit, sbe.HasIncompatibleFeature(bit))
	}
}

func ParseSuperblock(r io.Reader) (sb *Superblock, sbe *SuperblockExtension, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	sb = new(Superblock)

	err = binary.Read(r, binary.LittleEndian, sb)
	log.PanicIf(err)

	if sb.SMagic != Ext4Magic {
		log.Panic(ErrNotExt4)
	}

	if sb.SRevLevel != SbRevlevelDynamicRev {
		return sb, nil, nil
	}

	sbe = new(SuperblockExtension)

	err = binary.Read(r, binary.LittleEndian, sbe)
	log.PanicIf(err)

	return sb, sbe, nil
}
