package ext4

import (
	"fmt"
	"io"
	"time"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	Ext4NdirBlocks = 12
	Ext4IndBlock   = Ext4NdirBlocks
	Ext4DindBlock  = (Ext4IndBlock + 1)
	Ext4TindBlock  = (Ext4DindBlock + 1)
	Ext4NBlocks    = (Ext4TindBlock + 1)
)

type InodeData struct {
	IMode       uint16 /* File mode */
	IUid        uint16 /* Low 16 bits of Owner Uid */
	ISizeLo     uint32 /* Size in bytes */
	IAtime      uint32 /* Access time */
	ICtime      uint32 /* Inode Change time */
	IMtime      uint32 /* Modification time */
	IDtime      uint32 /* Deletion Time */
	IGid        uint16 /* Low 16 bits of Group Id */
	ILinksCount uint16 /* Links count */
	IBlocksLo   uint32 /* Blocks count */
	IFlags      uint32 /* File flags */

	// union {
	//     struct {
	//         __le32  l_i_version;
	//     } linux1;
	//     struct {
	//         __u32  h_i_translator;
	//     } hurd1;
	//     struct {
	//         __u32  m_i_reserved1;
	//     } masix1;
	// } osd1;             /* OS dependent 1 */
	Osd1 [4]byte

	IBlock      [Ext4NBlocks]uint32 /* Pointers to blocks */
	IGeneration uint32              /* File version (for NFS) */
	IFileAclLo  uint32              /* File ACL */
	ISizeHigh   uint32
	IObsoFaddr  uint32 /* Obsoleted fragment address */

	// union {
	//     struct {
	//         __le16  l_i_blocks_high; /* were l_i_reserved1 */
	//         __le16  l_i_file_acl_high;
	//         __le16  l_i_uid_high;   /* these 2 fields */
	//         __le16  l_i_gid_high;   /* were reserved2[0] */
	//         __le16  l_i_checksum_lo;/* crc32c(uuid+inum+inode) LE */
	//         __le16  l_i_reserved;
	//     } linux2;
	//     struct {
	//         __le16  h_i_reserved1;   Obsoleted fragment number/size which are removed in ext4
	//         __u16   h_i_mode_high;
	//         __u16   h_i_uid_high;
	//         __u16   h_i_gid_high;
	//         __u32   h_i_author;
	//     } hurd2;
	//     struct {
	//         __le16  h_i_reserved1;  /* Obsoleted fragment number/size which are removed in ext4 */
	//         __le16  m_i_file_acl_high;
	//         __u32   m_i_reserved2[2];
	//     } masix2;
	// } osd2;             /* OS dependent 2 */
	Osd2 [12]byte

	IExtraIsize  uint16
	IChecksumHi  uint16 /* crc32c(uuid+inum+inode) BE */
	ICtimeExtra  uint32 /* extra Change time      (nsec << 2 | epoch) */
	IMtimeExtra  uint32 /* extra Modification time(nsec << 2 | epoch) */
	IAtimeExtra  uint32 /* extra Access time      (nsec << 2 | epoch) */
	ICrtime      uint32 /* File Creation time */
	ICrtimeExtra uint32 /* extra FileCreationtime (nsec << 2 | epoch) */
	IVersionHi   uint32 /* high 32 bits for 64-bit version */
	IProjid      uint32 /* Project ID */
}

type Inode struct {
	data *InodeData
}

func NewInodeWithReadSeeker(bgd *BlockGroupDescriptor, rs io.ReadSeeker, absoluteInodeNumber int) (inode *Inode, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// TODO(dustin): !! We might want to find a way to verify this against the bitmap if we can pre-store it in the BGD.
	// func (bgd *BlockGroupDescriptor) InodeBitmapBlock() uint64 {

	sb := bgd.Superblock()

	absoluteInodeTableBlock := bgd.InodeTableBlock()
	offset := uint64(sb.BlockSize()) * absoluteInodeTableBlock

	// bgRelativeInode is the number of the inode within the inode-table for
	// this particular block-group. The math only makes sense if we take
	// (inode - 1) since there is no "inode 0".
	bgRelativeInode := (uint64(absoluteInodeNumber) - 1) % uint64(sb.Data().SInodesPerGroup)

	offset += bgRelativeInode * uint64(sb.Data().SInodeSize)

	_, err = rs.Seek(int64(offset), io.SeekStart)
	log.PanicIf(err)

	id := new(InodeData)

	err = binary.Read(rs, binary.LittleEndian, id)
	log.PanicIf(err)

	inode = &Inode{
		data: id,
	}

	return inode, nil
}

func (inode *Inode) Data() *InodeData {
	return inode.data
}

func (inode *Inode) AccessTime() time.Time {
	return time.Unix(int64(inode.Data().IAtime), 0)
}

func (inode *Inode) InodeChangeTime() time.Time {
	return time.Unix(int64(inode.Data().ICtime), 0)
}

func (inode *Inode) ModificationTime() time.Time {
	return time.Unix(int64(inode.Data().IMtime), 0)
}

func (inode *Inode) DeletionTime() time.Time {
	return time.Unix(int64(inode.Data().IDtime), 0)
}

func (inode *Inode) FileCreationTime() time.Time {
	return time.Unix(int64(inode.Data().ICrtime), 0)
}

func (inode *Inode) Dump() {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	fmt.Printf("IAtime: [%s]\n", inode.AccessTime())
	fmt.Printf("ICtime: [%s]\n", inode.InodeChangeTime())
	fmt.Printf("IMtime: [%s]\n", inode.ModificationTime())
	fmt.Printf("IDtime: [%s]\n", inode.DeletionTime())
	fmt.Printf("ICrtime: [%s]\n", inode.FileCreationTime())

	// TODO(dustin): !! Print the rest of the fields.

}
