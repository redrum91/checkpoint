package disk

import (
	"strings"
	"fmt"
)

// DriveGroup represents a logical grouping of disks
type DriveGroup struct {
	Name        string
	Icon        string
	Type        string
	TotalSize   uint64
	TotalUsed   uint64
	Available   uint64
	Disks       []Disk
	IsPrimary   bool
	Description string
}

// GroupDisks groups disks into logical drives for user-friendly display
func GroupDisks(disks []Disk) []DriveGroup {
	groups := []DriveGroup{}
	
	// First, find system drive (root partition)
	var systemGroup *DriveGroup
	var dataGroups []DriveGroup
	var removableGroups []DriveGroup
	
	// Track which disks have been grouped
	grouped := make(map[string]bool)
	
	// Find and create system drive group
	for _, disk := range disks {
		if disk.MountPoint == "/" {
			systemGroup = &DriveGroup{
				Name:        "System Drive",
				Icon:        "💻",
				Type:        "system",
				TotalSize:   disk.Size,
				TotalUsed:   disk.Used,
				Available:   disk.Available,
				Disks:       []Disk{disk},
				IsPrimary:   true,
				Description: "Linux System",
			}
			grouped[disk.Path] = true
			
			// Also add boot partitions to system group
			for _, d := range disks {
				if strings.HasPrefix(d.MountPoint, "/boot") {
					systemGroup.Disks = append(systemGroup.Disks, d)
					systemGroup.TotalSize += d.Size
					systemGroup.TotalUsed += d.Used
					grouped[d.Path] = true
				}
			}
			break
		}
	}
	
	// Group data drives by physical disk
	physicalDisks := make(map[string]*DriveGroup)
	
	for _, disk := range disks {
		// Skip if already grouped, loops, or system mounts
		if grouped[disk.Path] || disk.Type == TypeLoop || 
		   strings.HasPrefix(disk.MountPoint, "/snap") ||
		   strings.HasPrefix(disk.MountPoint, "/run") ||
		   strings.HasPrefix(disk.MountPoint, "/sys") ||
		   strings.HasPrefix(disk.MountPoint, "/proc") {
			continue
		}
		
		// Skip tiny partitions (like EFI)
		if disk.Size < 1024*1024*1024 { // Less than 1GB
			continue
		}
		
		// Determine base device name
		baseName := getBaseDiskName(disk.Path)
		
		// Create or update group
		if group, exists := physicalDisks[baseName]; exists {
			group.Disks = append(group.Disks, disk)
			group.TotalSize += disk.Size
			group.TotalUsed += disk.Used
			group.Available += disk.Available
		} else {
			name := getDriveName(disk, len(dataGroups)+1)
			physicalDisks[baseName] = &DriveGroup{
				Name:        name,
				Icon:        getDriveIcon(disk),
				Type:        "data",
				TotalSize:   disk.Size,
				TotalUsed:   disk.Used,
				Available:   disk.Available,
				Disks:       []Disk{disk},
				IsPrimary:   false,
				Description: getDriveDescription(disk),
			}
		}
		grouped[disk.Path] = true
	}
	
	// Convert map to slice
	for _, group := range physicalDisks {
		dataGroups = append(dataGroups, *group)
	}
	
	// Add removable/network drives
	for _, disk := range disks {
		if grouped[disk.Path] {
			continue
		}
		
		if disk.Type == TypeNetwork || disk.Type == TypeFUSE {
			removableGroups = append(removableGroups, DriveGroup{
				Name:        getNetworkDriveName(disk),
				Icon:        "🌐",
				Type:        "network",
				TotalSize:   disk.Size,
				TotalUsed:   disk.Used,
				Available:   disk.Available,
				Disks:       []Disk{disk},
				IsPrimary:   false,
				Description: "Network Storage",
			})
		}
	}
	
	// Assemble final groups list
	if systemGroup != nil {
		groups = append(groups, *systemGroup)
	}
	groups = append(groups, dataGroups...)
	groups = append(groups, removableGroups...)
	
	return groups
}

// getBaseDiskName extracts base disk name (e.g., /dev/sda from /dev/sda1)
func getBaseDiskName(path string) string {
	// Remove partition numbers
	base := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] >= '0' && path[i] <= '9' {
			base = path[:i]
		} else {
			break
		}
	}
	return base
}

// getDriveName generates a friendly name for the drive
func getDriveName(disk Disk, index int) string {
	// Check mount point for hints
	switch {
	case strings.Contains(disk.MountPoint, "home"):
		return "Home Drive"
	case strings.Contains(disk.MountPoint, "data"):
		return "Data Drive"
	case strings.Contains(disk.MountPoint, "backup"):
		return "Backup Drive"
	case strings.Contains(disk.MountPoint, "media"):
		return "Media Drive"
	case disk.Type == TypeLVM:
		return fmt.Sprintf("Volume %d", index)
	case strings.HasPrefix(disk.Path, "/dev/nvme"):
		return fmt.Sprintf("SSD Drive %d", index)
	case strings.HasPrefix(disk.Path, "/dev/sd"):
		return fmt.Sprintf("Drive %d", index)
	default:
		return fmt.Sprintf("Storage %d", index)
	}
}

// getDriveIcon returns appropriate icon for drive type
func getDriveIcon(disk Disk) string {
	switch {
	case strings.HasPrefix(disk.Path, "/dev/nvme"):
		return "⚡" // SSD
	case disk.Type == TypeNetwork:
		return "🌐"
	case disk.Type == TypeLVM:
		return "🗄️"
	default:
		return "💾" // HDD
	}
}

// getDriveDescription generates a description for the drive
func getDriveDescription(disk Disk) string {
	switch {
	case strings.HasPrefix(disk.Path, "/dev/nvme"):
		return "NVMe SSD"
	case disk.Type == TypeLVM:
		return "Logical Volume"
	case strings.HasPrefix(disk.Path, "/dev/sd"):
		return "Hard Drive"
	default:
		return "Storage Device"
	}
}

// getNetworkDriveName generates name for network drives
func getNetworkDriveName(disk Disk) string {
	if disk.Type == TypeNetwork {
		// Extract server name if possible
		parts := strings.Split(disk.Path, "/")
		if len(parts) > 2 {
			return fmt.Sprintf("Network (%s)", parts[2])
		}
		return "Network Drive"
	}
	return "Remote Storage"
}