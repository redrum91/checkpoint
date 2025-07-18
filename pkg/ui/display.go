package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"checkpoint/pkg/disk"
)

var (
	// Define styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginTop(1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("241"))

	rowStyle = lipgloss.NewStyle().
			PaddingRight(2)

	evenRowStyle = rowStyle.Copy().
			Background(lipgloss.Color("235"))

	diskTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	symlinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Italic(true)

	hardlinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213"))

	sizeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	availableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	usedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	inodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	totalStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("241")).
			MarginTop(1).
			PaddingTop(1)

	legendStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			MarginTop(1)

	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			MarginTop(1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
)

func DisplayDisks(disks []disk.Disk, showDetails bool) {
	title := titleStyle.Render("💾 Storage Disks Overview")
	fmt.Println(title)

	if showDetails {
		displayDetailedView(disks)
	} else {
		displaySimpleView(disks)
	}
}

func displaySimpleView(disks []disk.Disk) {
	// Simple view - no inode column, condensed display
	headers := []string{"ID", "Device", "Type", "Size", "Available", "Mount"}
	headerRow := makeSimpleRow(headers, headerStyle)
	fmt.Println(headerRow)

	// Display only physical and important disks
	displayCount := 0
	for _, d := range disks {
		// Skip symlinks in simple view
		if d.Type == disk.TypeSymlink {
			continue
		}
		
		displayCount++
		rowData := formatSimpleDiskRow(displayCount, d)
		style := rowStyle
		if displayCount%2 == 0 {
			style = evenRowStyle
		}
		fmt.Println(makeSimpleStyledRow(rowData, style))
	}
}

func displayDetailedView(disks []disk.Disk) {
	// Detailed view with all information
	headers := []string{"ID", "Device", "Type", "FS", "Size", "Used", "Available", "Inode", "Mount"}
	headerRow := makeRow(headers, headerStyle)
	fmt.Println(headerRow)

	hardlinks := make(map[uint64][]int)
	for i, d := range disks {
		if d.Inode > 0 && d.Type != disk.TypeSymlink {
			hardlinks[d.Inode] = append(hardlinks[d.Inode], i)
		}
	}

	for i, d := range disks {
		rowData := formatDiskRow(i+1, d, hardlinks)
		style := rowStyle
		if i%2 == 0 {
			style = evenRowStyle
		}
		fmt.Println(makeStyledRow(rowData, style))
	}
}

func formatSimpleDiskRow(id int, d disk.Disk) []string {
	devicePath := truncatePath(d.Path, 30)
	typeStr := string(d.Type)
	
	return []string{
		fmt.Sprintf("%d", id),
		devicePath,
		diskTypeStyle.Render(typeStr),
		sizeStyle.Render(FormatBytes(d.Size)),
		availableStyle.Render(FormatBytes(d.Available)),
		truncatePath(d.MountPoint, 40),
	}
}

func makeSimpleRow(cols []string, style lipgloss.Style) string {
	widths := []int{4, 32, 12, 12, 12, 42}
	return makeRowWithWidths(cols, style, widths)
}

func makeSimpleStyledRow(cols []string, baseStyle lipgloss.Style) string {
	widths := []int{4, 32, 12, 12, 12, 42}
	return makeStyledRowWithWidths(cols, baseStyle, widths)
}

func makeRowWithWidths(cols []string, style lipgloss.Style, widths []int) string {
	styledCols := make([]string, len(cols))
	for i, col := range cols {
		if i < len(widths) {
			styledCols[i] = style.Render(lipgloss.NewStyle().Width(widths[i]).Render(col))
		} else {
			styledCols[i] = style.Render(col)
		}
	}
	return strings.Join(styledCols, " ")
}

func makeStyledRowWithWidths(cols []string, baseStyle lipgloss.Style, widths []int) string {
	styledCols := make([]string, len(cols))
	for i, col := range cols {
		if i < len(widths) {
			padded := lipgloss.NewStyle().Width(widths[i]).Render(col)
			styledCols[i] = baseStyle.Render(padded)
		} else {
			styledCols[i] = baseStyle.Render(col)
		}
	}
	return strings.Join(styledCols, " ")
}

func formatDiskRow(id int, d disk.Disk, hardlinks map[uint64][]int) []string {
	// Format device path
	devicePath := d.Path
	if d.IsSymlink && d.LinkTarget != "" {
		devicePath = fmt.Sprintf("%s → %s", truncatePath(d.Path, 20), truncatePath(d.LinkTarget, 20))
		devicePath = symlinkStyle.Render(devicePath)
	} else {
		devicePath = truncatePath(devicePath, 40)
	}

	// Format type with icons
	typeStr := formatDiskType(d.Type)
	if d.IsSymlink {
		typeStr = "✨ " + typeStr
	}
	if links, ok := hardlinks[d.Inode]; ok && len(links) > 1 && d.Type != disk.TypeSymlink {
		typeStr = "🔗 " + typeStr
		typeStr = hardlinkStyle.Render(typeStr)
	} else {
		typeStr = diskTypeStyle.Render(typeStr)
	}

	// Format inode
	inodeStr := "-"
	if d.Inode > 0 {
		inodeStr = fmt.Sprintf("%d", d.Inode)
	}
	inodeStr = inodeStyle.Render(inodeStr)

	return []string{
		fmt.Sprintf("%d", id),
		devicePath,
		typeStr,
		d.Filesystem,
		sizeStyle.Render(FormatBytes(d.Size)),
		usedStyle.Render(FormatBytes(d.Used)),
		availableStyle.Render(FormatBytes(d.Available)),
		inodeStr,
		truncatePath(d.MountPoint, 30),
	}
}

func formatDiskType(t disk.DiskType) string {
	icons := map[disk.DiskType]string{
		disk.TypePhysical: "💽",
		disk.TypeLVM:      "🗄️",
		disk.TypeLoop:     "🔄",
		disk.TypeBind:     "📁",
		disk.TypeNetwork:  "🌐",
		disk.TypeFUSE:     "🔌",
		disk.TypePath:     "📂",
		disk.TypeManual:   "✋",
		disk.TypeSymlink:  "🔗",
	}

	icon, ok := icons[t]
	if !ok {
		icon = "❓"
	}
	return fmt.Sprintf("%s %s", icon, t)
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen < 10 {
		return path[:maxLen]
	}
	return "..." + path[len(path)-(maxLen-3):]
}


func makeRow(cols []string, style lipgloss.Style) string {
	widths := []int{4, 42, 15, 10, 12, 12, 12, 10, 32}
	styledCols := make([]string, len(cols))
	
	for i, col := range cols {
		if i < len(widths) {
			styledCols[i] = style.Render(lipgloss.NewStyle().Width(widths[i]).Render(col))
		} else {
			styledCols[i] = style.Render(col)
		}
	}
	
	return strings.Join(styledCols, " ")
}

func makeStyledRow(cols []string, baseStyle lipgloss.Style) string {
	widths := []int{4, 42, 15, 10, 12, 12, 12, 10, 32}
	styledCols := make([]string, len(cols))
	
	for i, col := range cols {
		if i < len(widths) {
			// Note: col might already be styled, so we just pad it
			padded := lipgloss.NewStyle().Width(widths[i]).Render(col)
			styledCols[i] = baseStyle.Render(padded)
		} else {
			styledCols[i] = baseStyle.Render(col)
		}
	}
	
	return strings.Join(styledCols, " ")
}

func DisplayMenu() {
	menu := menuStyle.Render("Options:") + "\n" +
		menuItemStyle.Render("1.") + " Add a disk path manually\n" +
		menuItemStyle.Render("2.") + " Execute installation command\n" +
		menuItemStyle.Render("3.") + " Rescan disks\n" +
		menuItemStyle.Render("4.") + " Exit"
	
	fmt.Println(menu)
	fmt.Print(promptStyle.Render("Select option: "))
}