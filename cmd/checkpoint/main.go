package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"checkpoint/pkg/disk"
	"checkpoint/pkg/installer"
	"checkpoint/pkg/ui"
)

var (
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
)

func main() {
	dm := disk.NewManager()
	scanner := bufio.NewScanner(os.Stdin)
	showDetails := false
	friendlyView := true // New default view

	// Initial scan
	if err := dm.ScanDisks(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error scanning disks: %v", err)))
	}

	for {
		// Clear screen for better display
		fmt.Print("\033[H\033[2J")
		
		// Display based on view mode
		if friendlyView {
			// Group disks for friendly view
			groups := disk.GroupDisks(dm.GetDisks())
			ui.DisplayFriendlyDisks(groups)
		} else {
			// Traditional view
			stats := dm.GetStats()
			ui.DisplaySummary(stats, dm.GetDisks())
			ui.DisplayDisks(dm.GetDisks(), showDetails)
		}
		
		// Enhanced menu
		displayEnhancedMenu(friendlyView)

		if !scanner.Scan() {
			break
		}

		option := strings.TrimSpace(scanner.Text())

		switch option {
		case "1":
			handleAddDisk(dm, scanner)
		case "2":
			handleInstallCommand(dm, scanner, friendlyView)
		case "3":
			handleRescan(dm)
		case "4":
			if friendlyView {
				friendlyView = false
				fmt.Println(infoStyle.Render("📊 Switched to technical view"))
			} else {
				showDetails = !showDetails
				fmt.Println(infoStyle.Render(fmt.Sprintf("📊 Detail view: %v", showDetails)))
			}
		case "5":
			friendlyView = !friendlyView
			fmt.Println(infoStyle.Render(fmt.Sprintf("🖥️ Friendly view: %v", friendlyView)))
		case "6":
			fmt.Println(infoStyle.Render("👋 Exiting..."))
			return
		default:
			fmt.Println(errorStyle.Render("❌ Invalid option"))
		}

		if option != "6" {
			fmt.Println(infoStyle.Render("\nPress Enter to continue..."))
			scanner.Scan()
		}
	}
}

func displayEnhancedMenu(friendlyView bool) {
	menu := infoStyle.Render("Options:") + "\n" +
		successStyle.Render("1.") + " Add a disk path manually\n" +
		successStyle.Render("2.") + " Execute installation command (no sudo required)\n" +
		successStyle.Render("3.") + " Rescan disks\n"
	
	if friendlyView {
		menu += successStyle.Render("4.") + " Switch to technical view\n"
	} else {
		menu += successStyle.Render("4.") + " Toggle detailed view\n"
	}
	
	menu += successStyle.Render("5.") + " Toggle view mode (friendly/technical)\n" +
		successStyle.Render("6.") + " Exit"
	
	fmt.Println(menu)
	fmt.Print(infoStyle.Render("Select option: "))
}

func handleAddDisk(dm *disk.Manager, scanner *bufio.Scanner) {
	// Check for unmounted disks
	unmounted, _ := disk.ScanUnmountedDisks()
	
	fmt.Println(infoStyle.Render("\n📁 Add Disk Path"))
	
	// Show suggestions
	if len(unmounted) > 0 {
		fmt.Println(infoStyle.Render("\n💿 Unmounted disks detected:"))
		for i, ud := range unmounted {
			fmt.Printf("%s%d.%s %s (%s, %s)\n", 
				successStyle.Render(fmt.Sprintf("%d", i+1)),
				successStyle.Render("."),
				ud.Device,
				ud.Size,
				ud.Filesystem)
			if ud.Label != "" {
				fmt.Printf("   Label: %s\n", ud.Label)
			}
		}
		fmt.Println(infoStyle.Render("\nNote: These disks need to be mounted first to be used"))
	}
	
	// Show directory suggestions
	dirs := disk.GetMountableDirectories()
	if len(dirs) > 0 {
		fmt.Println(infoStyle.Render("\n📂 Suggested directories:"))
		for i, dir := range dirs {
			if i < 5 { // Show max 5 suggestions
				fmt.Printf("  • %s\n", dir)
			}
		}
	}
	
	fmt.Print(infoStyle.Render("\n📁 Enter disk path (or press Enter to cancel): "))
	if scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path == "" {
			fmt.Println(infoStyle.Render("❌ Cancelled - no disk added"))
			return
		}
		
		if err := dm.AddCustomPath(path); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error adding disk: %v", err)))
		} else {
			fmt.Println(successStyle.Render("✅ Disk added successfully"))
		}
	}
}

func handleInstallCommand(dm *disk.Manager, scanner *bufio.Scanner, friendlyView bool) {
	// Show package manager info
	pm := installer.DetectPackageManager()
	if pm != "unknown" {
		fmt.Println(infoStyle.Render(fmt.Sprintf("📦 Detected package manager: %s", pm)))
	}

	fmt.Print(infoStyle.Render("💻 Enter installation command: "))
	if !scanner.Scan() {
		return
	}
	
	command := strings.TrimSpace(scanner.Text())
	if command == "" {
		fmt.Println(errorStyle.Render("❌ Empty command"))
		return
	}

	// Show drive selection based on view
	var targetDisk *disk.Disk
	
	if friendlyView {
		// Show friendly drive groups
		groups := disk.GroupDisks(dm.GetDisks())
		fmt.Println(infoStyle.Render("\n🎯 Select target drive:"))
		for i, group := range groups {
			fmt.Printf("%s. %s %s (%s free)\n", 
				successStyle.Render(fmt.Sprintf("%d", i+1)),
				group.Icon,
				group.Name,
				ui.FormatBytes(group.Available))
		}
		fmt.Print(infoStyle.Render("Select drive (or press Enter for default): "))
		
		if scanner.Scan() {
			driveIDStr := strings.TrimSpace(scanner.Text())
			if driveIDStr != "" {
				driveID, err := strconv.Atoi(driveIDStr)
				if err == nil && driveID > 0 && driveID <= len(groups) {
					// Use the first disk in the selected group
					if len(groups[driveID-1].Disks) > 0 {
						targetDisk = &groups[driveID-1].Disks[0]
					}
				}
			}
		}
	} else {
		// Traditional disk selection
		fmt.Print(infoStyle.Render("🎯 Select target disk ID (or press Enter for default): "))
		if scanner.Scan() {
			diskIDStr := strings.TrimSpace(scanner.Text())
			if diskIDStr != "" {
				diskID, err := strconv.Atoi(diskIDStr)
				if err != nil || diskID < 1 || diskID > len(dm.GetDisks()) {
					fmt.Println(errorStyle.Render("❌ Invalid disk ID"))
					return
				}
				disks := dm.GetDisks()
				targetDisk = &disks[diskID-1]
			}
		}
	}

	// Execute command
	if err := installer.ExecuteCommand(command, targetDisk); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error executing command: %v", err)))
	} else {
		fmt.Println(successStyle.Render("✅ Command executed successfully"))
	}
}

func handleRescan(dm *disk.Manager) {
	fmt.Println(infoStyle.Render("🔄 Rescanning disks..."))
	dm.ClearDisks()
	if err := dm.ScanDisks(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error rescanning disks: %v", err)))
	} else {
		fmt.Println(successStyle.Render("✅ Rescan completed"))
	}
}