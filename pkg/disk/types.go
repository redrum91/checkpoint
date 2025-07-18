package disk

import "time"

type Disk struct {
	Path       string
	Filesystem string
	Size       uint64
	Available  uint64
	Used       uint64
	MountPoint string
	Type       DiskType
	IsSymlink  bool
	LinkTarget string
	Inode      uint64
	Device     string
	LastCheck  time.Time
}

type DiskType string

const (
	TypePhysical DiskType = "physical"
	TypeLVM      DiskType = "lvm"
	TypeLoop     DiskType = "loop"
	TypeBind     DiskType = "bind"
	TypeNetwork  DiskType = "network"
	TypeFUSE     DiskType = "fuse"
	TypePath     DiskType = "path"
	TypeManual   DiskType = "manual"
	TypeSymlink  DiskType = "symlink"
)

type Manager struct {
	disks     []Disk
	lastScan  time.Time
	scanCache map[string]bool
}

func NewManager() *Manager {
	return &Manager{
		disks:     make([]Disk, 0),
		scanCache: make(map[string]bool),
	}
}

func (m *Manager) GetDisks() []Disk {
	return m.disks
}

func (m *Manager) AddDisk(disk Disk) {
	m.disks = append(m.disks, disk)
}

func (m *Manager) ClearDisks() {
	m.disks = make([]Disk, 0)
	m.scanCache = make(map[string]bool)
}