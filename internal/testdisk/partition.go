package testdisk

import (
	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/disk"
)

const FakePartitionSize = uint64(789) * common.MiB

// MakeFakePartitionTable is a helper to create partition table structs
// for tests. It uses sensible defaults for common scenarios.
func MakeFakePartitionTable(mntPoints ...string) *disk.PartitionTable {
	var partitions []disk.Partition
	for _, mntPoint := range mntPoints {
		payload := &disk.Filesystem{
			Type:       "ext4",
			Mountpoint: mntPoint,
		}
		switch mntPoint {
		case "/":
			payload.UUID = disk.RootPartitionUUID
		case "/boot/efi":
			payload.UUID = disk.EFIFilesystemUUID
			payload.Type = "vfat"
		default:
			payload.UUID = disk.FilesystemDataUUID
		}
		partitions = append(partitions, disk.Partition{
			Size:    FakePartitionSize,
			Payload: payload,
		})

	}
	return &disk.PartitionTable{
		Type:       "gpt",
		Partitions: partitions,
	}
}

// MakeFakeBtrfsPartitionTable is similar to MakeFakePartitionTable but
// creates a btrfs-based partition table.
func MakeFakeBtrfsPartitionTable(mntPoints ...string) *disk.PartitionTable {
	var subvolumes []disk.BtrfsSubvolume
	pt := &disk.PartitionTable{
		Type:       "gpt",
		Size:       10 * common.GiB,
		Partitions: []disk.Partition{},
	}
	size := uint64(0)
	for _, mntPoint := range mntPoints {
		switch mntPoint {
		case "/boot":
			pt.Partitions = append(pt.Partitions, disk.Partition{
				Start: size,
				Size:  1 * common.GiB,
				Payload: &disk.Filesystem{
					Type:       "ext4",
					Mountpoint: mntPoint,
				},
			})
			size += 1 * common.GiB
		case "/boot/efi":
			pt.Partitions = append(pt.Partitions, disk.Partition{
				Start: size,
				Size:  100 * common.MiB,
				Payload: &disk.Filesystem{
					Type:       "vfat",
					Mountpoint: mntPoint,
					UUID:       disk.EFIFilesystemUUID,
				},
			})
			size += 100 * common.MiB
		default:
			name := mntPoint
			if name == "/" {
				name = "root"
			}
			subvolumes = append(
				subvolumes,
				disk.BtrfsSubvolume{
					Mountpoint: mntPoint,
					Name:       name,
					UUID:       disk.RootPartitionUUID,
					Compress:   "zstd:1",
				},
			)
		}
	}

	pt.Partitions = append(pt.Partitions, disk.Partition{
		Start: size,
		Size:  9 * common.GiB,
		Payload: &disk.Btrfs{
			UUID:       disk.RootPartitionUUID,
			Subvolumes: subvolumes,
		},
	})

	size += 9 * common.GiB
	pt.Size = size

	return pt
}
