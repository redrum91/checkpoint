package disk

import (
	"fmt"
	"sort"
)

// DiskStats holds statistics about the disks
type DiskStats struct {
	TotalDisks     int
	TotalSize      uint64
	TotalAvailable uint64
	TotalUsed      uint64
	DisksByType    map[DiskType]int
	Hardlinks      map[uint64][]string // inode -> paths
	Symlinks       []SymlinkInfo
}

type SymlinkInfo struct {
	Source string
	Target string
}

// GetStats analyzes disks and returns statistics
func (m *Manager) GetStats() DiskStats {
	stats := DiskStats{
		TotalDisks:  len(m.disks),
		DisksByType: make(map[DiskType]int),
		Hardlinks:   make(map[uint64][]string),
		Symlinks:    make([]SymlinkInfo, 0),
	}

	// Analyze each disk
	for _, disk := range m.disks {
		// Count by type
		stats.DisksByType[disk.Type]++

		// Skip virtual/special disks for size calculations
		if disk.Type != TypeSymlink {
			stats.TotalSize += disk.Size
			stats.TotalAvailable += disk.Available
			stats.TotalUsed += disk.Used
		}

		// Track hardlinks (multiple paths with same inode)
		if disk.Inode > 0 && disk.Type != TypeSymlink {
			stats.Hardlinks[disk.Inode] = append(stats.Hardlinks[disk.Inode], disk.Path)
		}

		// Track symlinks
		if disk.IsSymlink && disk.LinkTarget != "" {
			stats.Symlinks = append(stats.Symlinks, SymlinkInfo{
				Source: disk.Path,
				Target: disk.LinkTarget,
			})
		}
	}

	// Clean up hardlinks - only keep those with multiple paths
	for inode, paths := range stats.Hardlinks {
		if len(paths) <= 1 {
			delete(stats.Hardlinks, inode)
		}
	}

	return stats
}

// GetSummary returns a human-readable summary
func (s DiskStats) GetSummary() string {
	summary := fmt.Sprintf("Storage Summary:\n")
	summary += fmt.Sprintf("• Total disks: %d\n", s.TotalDisks)
	summary += fmt.Sprintf("• Total capacity: %s\n", formatBytes(s.TotalSize))
	summary += fmt.Sprintf("• Used: %s (%.1f%%)\n", formatBytes(s.TotalUsed), float64(s.TotalUsed)/float64(s.TotalSize)*100)
	summary += fmt.Sprintf("• Available: %s\n", formatBytes(s.TotalAvailable))

	if len(s.DisksByType) > 0 {
		summary += "\nDisk types:\n"
		types := make([]DiskType, 0, len(s.DisksByType))
		for t := range s.DisksByType {
			types = append(types, t)
		}
		sort.Slice(types, func(i, j int) bool {
			return string(types[i]) < string(types[j])
		})
		for _, t := range types {
			summary += fmt.Sprintf("• %s: %d\n", t, s.DisksByType[t])
		}
	}

	if len(s.Hardlinks) > 0 {
		summary += fmt.Sprintf("\nHard links detected: %d groups\n", len(s.Hardlinks))
	}

	if len(s.Symlinks) > 0 {
		summary += fmt.Sprintf("Symbolic links: %d\n", len(s.Symlinks))
	}

	return summary
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}