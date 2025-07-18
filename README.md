# Checkpoint - Linux Disk Manager CLI

## Introduction

This project was created for newcomers like myself transitioning from Windows to Linux, to help understand the Linux filesystem and make storage management and software installation simpler and more intuitive. It bridges the gap between Windows' familiar drive interface and Linux's powerful but sometimes complex storage system.

## Features

- 🖥️ **Windows-like View**: Friendly interface that groups drives similar to "My Computer" 
- 📊 **Smart Grouping**: Automatically groups system partitions and hides technical details
- 🎨 **Visual Progress Bars**: Clear disk usage visualization with colored progress bars
- 🚀 **Safe Installation**: Execute commands without sudo/root requirements
- 📦 **Auto-detection**: Detects unmounted disks and suggests mount points
- 🔄 **Dual Views**: Switch between friendly (Windows-like) and technical (traditional Linux) views

## Installation

Build from source:

```bash
git clone https://github.com/redrum91/checkpoint.git
cd checkpoint
go build -o checkpoint ./cmd/checkpoint
```

## Usage

```bash
./checkpoint
```

The app starts in friendly view by default, showing drives in a Windows-like format.

### Menu Options

1. **Add disk path** - Manually add directories or see unmounted disks
2. **Execute installation** - Run commands with visual drive selection
3. **Rescan disks** - Refresh the disk list
4. **Switch views** - Toggle between friendly/technical views
5. **Toggle view mode** - Quick switch between view types
6. **Exit** - Quit the application

### Views

**Friendly View (Default)**
- Groups system partitions together
- Shows drives with descriptive names
- Visual progress bars for disk usage
- Hides technical details (loop devices, etc.)

**Technical View**
- Traditional Linux disk listing
- Shows all mount points and devices
- Detailed information with inodes
- Toggle between simple/detailed display

## Requirements

- Go 1.24 or later
- No root/sudo required - runs with user permissions

## Contributing

This project was built quickly to address an immediate need for Windows users transitioning to Linux. While functional, there's plenty of room for improvements and new features. Contributions are welcome!

Feel free to:
- Report bugs or suggest features via issues
- Submit pull requests with improvements
- Share your ideas for making Linux more accessible

## Disclaimer

Just in case there are any issues due to this program, I am not responsible. Use at your own risk.