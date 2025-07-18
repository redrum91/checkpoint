package disk

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// UnmountedDisk represents a disk that is not currently mounted
type UnmountedDisk struct {
	Device     string
	Size       string
	Type       string
	Label      string
	UUID       string
	Filesystem string
}

// ScanUnmountedDisks finds disks that are not currently mounted
func ScanUnmountedDisks() ([]UnmountedDisk, error) {
	unmounted := []UnmountedDisk{}
	
	// Get all block devices using lsblk
	cmd := exec.Command("lsblk", "-rno", "NAME,SIZE,TYPE,LABEL,UUID,FSTYPE")
	output, err := cmd.Output()
	if err != nil {
		// If lsblk is not available, return empty list
		return unmounted, nil
	}

	// Get list of mounted devices
	mounted := make(map[string]bool)
	mounts, _ := os.ReadFile("/proc/mounts")
	for _, line := range strings.Split(string(mounts), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			device := fields[0]
			mounted[device] = true
			// Also mark without /dev/ prefix
			if strings.HasPrefix(device, "/dev/") {
				mounted[strings.TrimPrefix(device, "/dev/")] = true
			}
		}
	}

	// Parse lsblk output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		name := fields[0]
		size := fields[1]
		diskType := fields[2]
		
		// Skip if not a disk or partition
		if diskType != "disk" && diskType != "part" {
			continue
		}

		// Check if device is mounted
		devicePath := "/dev/" + name
		if mounted[devicePath] || mounted[name] {
			continue
		}

		// Extract additional info
		label := ""
		uuid := ""
		fstype := ""
		
		if len(fields) > 3 && fields[3] != "" {
			label = fields[3]
		}
		if len(fields) > 4 && fields[4] != "" {
			uuid = fields[4]
		}
		if len(fields) > 5 && fields[5] != "" {
			fstype = fields[5]
		}

		// Skip if no filesystem
		if fstype == "" {
			continue
		}

		unmounted = append(unmounted, UnmountedDisk{
			Device:     devicePath,
			Size:       size,
			Type:       diskType,
			Label:      label,
			UUID:       uuid,
			Filesystem: fstype,
		})
	}

	return unmounted, nil
}

// GetMountableDirectories returns directories that could be mount points
func GetMountableDirectories() []string {
	suggestions := []string{}
	
	// Check common mount point directories
	checkDirs := []string{"/mnt", "/media", "/run/media", os.Getenv("HOME") + "/mnt"}
	
	for _, dir := range checkDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// List subdirectories
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			
			for _, entry := range entries {
				if entry.IsDir() {
					fullPath := fmt.Sprintf("%s/%s", dir, entry.Name())
					// Check if it's empty (potential mount point)
					subEntries, _ := os.ReadDir(fullPath)
					if len(subEntries) == 0 {
						suggestions = append(suggestions, fullPath)
					}
				}
			}
			
			// Also suggest the base directory if writable
			if canWrite(dir) {
				suggestions = append(suggestions, dir)
			}
		}
	}
	
	// Add user's home directory subdirs
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		suggestions = append(suggestions, homeDir+"/Downloads")
		suggestions = append(suggestions, homeDir+"/Documents")
	}
	
	return suggestions
}

func canWrite(path string) bool {
	testFile := fmt.Sprintf("%s/.checkpoint_test_%d", path, os.Getpid())
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}