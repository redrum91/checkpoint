package disk

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	// Generic Linux paths - no hardcoded specific paths
	defaultScanPaths = []string{
		"/",
		"/usr",
		"/opt",
		"/home",
		"/var",
		"/srv",
		"/media",
		"/mnt",
	}
	
	// Virtual filesystems to skip
	virtualFS = map[string]bool{
		"tmpfs":       true,
		"devtmpfs":    true,
		"sysfs":       true,
		"proc":        true,
		"cgroup":      true,
		"cgroup2":     true,
		"debugfs":     true,
		"securityfs":  true,
		"pstore":      true,
		"efivarfs":    true,
		"bpf":         true,
		"tracefs":     true,
		"hugetlbfs":   true,
		"mqueue":      true,
		"configfs":    true,
		"ramfs":       true,
		"autofs":      true,
		"fusectl":     true,
	}
)

func (m *Manager) ScanDisks() error {
	m.lastScan = time.Now()
	
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("failed to open /proc/mounts: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	seenMounts := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		filesystem := fields[2]
		options := fields[3]

		if seenMounts[mountPoint] {
			continue
		}
		seenMounts[mountPoint] = true

		if virtualFS[filesystem] {
			continue
		}

		disk := m.analyzeDisk(device, mountPoint, filesystem, options)
		if disk != nil {
			m.disks = append(m.disks, *disk)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading /proc/mounts: %v", err)
	}

	// Scan for symbolic links after main scan
	m.scanSymlinks()
	
	return nil
}

func (m *Manager) analyzeDisk(device, mountPoint, filesystem, options string) *Disk {
	diskType := determineDiskType(device, filesystem, options)
	if diskType == "" {
		return nil
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &stat); err != nil {
		return nil
	}

	disk := &Disk{
		Path:       device,
		Device:     device,
		Filesystem: filesystem,
		Size:       stat.Blocks * uint64(stat.Bsize),
		Available:  stat.Bavail * uint64(stat.Bsize),
		Used:       (stat.Blocks - stat.Bfree) * uint64(stat.Bsize),
		MountPoint: mountPoint,
		Type:       diskType,
		LastCheck:  time.Now(),
	}

	// Check if device is a symlink
	if info, err := os.Lstat(device); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			disk.IsSymlink = true
			if target, err := filepath.EvalSymlinks(device); err == nil {
				disk.LinkTarget = target
			}
		}
		if sysStat, ok := info.Sys().(*syscall.Stat_t); ok {
			disk.Inode = sysStat.Ino
		}
	}

	// Get inode of mount point if device inode failed
	if disk.Inode == 0 {
		if info, err := os.Stat(mountPoint); err == nil {
			if sysStat, ok := info.Sys().(*syscall.Stat_t); ok {
				disk.Inode = sysStat.Ino
			}
		}
	}

	return disk
}

func determineDiskType(device, filesystem, options string) DiskType {
	switch {
	case strings.Contains(options, "bind"):
		return TypeBind
	case strings.HasPrefix(device, "/dev/loop"):
		return TypeLoop
	case strings.HasPrefix(device, "/dev/mapper/"):
		return TypeLVM
	case strings.HasPrefix(device, "/dev/"):
		return TypePhysical
	case filesystem == "nfs" || filesystem == "nfs4" || filesystem == "cifs" || filesystem == "smb":
		return TypeNetwork
	case filesystem == "fuse" || strings.Contains(filesystem, "fuse"):
		return TypeFUSE
	case strings.HasPrefix(device, "/") && !strings.HasPrefix(device, "/dev/"):
		return TypePath
	default:
		return ""
	}
}

func (m *Manager) scanSymlinks() {
	// Skip symlink scanning by default - too many results
	// This can be enabled with a flag if needed
	return
}

func (m *Manager) walkDir(path string, currentDepth, maxDepth int, maxPathLen int) {
	if currentDepth >= maxDepth || len(path) > maxPathLen {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		
		// Skip if we've seen this path
		if m.scanCache[fullPath] {
			continue
		}
		m.scanCache[fullPath] = true

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if target, err := filepath.EvalSymlinks(fullPath); err == nil {
				if stat, err := os.Stat(target); err == nil && stat.IsDir() {
					// Check if symlink points to a mounted directory
					for _, disk := range m.disks {
						if strings.HasPrefix(target, disk.MountPoint) && target != disk.MountPoint {
							symlinkDisk := Disk{
								Path:       fullPath,
								Device:     fullPath,
								Filesystem: disk.Filesystem,
								MountPoint: fullPath,
								Type:       TypeSymlink,
								IsSymlink:  true,
								LinkTarget: target,
								LastCheck:  time.Now(),
							}
							m.disks = append(m.disks, symlinkDisk)
							break
						}
					}
				}
			}
		} else if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			m.walkDir(fullPath, currentDepth+1, maxDepth, maxPathLen)
		}
	}
}

func (m *Manager) AddCustomPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat path: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(absPath, &stat); err != nil {
		return fmt.Errorf("failed to get filesystem stats: %v", err)
	}

	var inode uint64
	if sysStat, ok := info.Sys().(*syscall.Stat_t); ok {
		inode = sysStat.Ino
	}

	disk := Disk{
		Path:       absPath,
		Device:     absPath,
		Filesystem: "unknown",
		Size:       stat.Blocks * uint64(stat.Bsize),
		Available:  stat.Bavail * uint64(stat.Bsize),
		Used:       (stat.Blocks - stat.Bfree) * uint64(stat.Bsize),
		MountPoint: absPath,
		Type:       TypeManual,
		Inode:      inode,
		LastCheck:  time.Now(),
	}

	m.disks = append(m.disks, disk)
	return nil
}