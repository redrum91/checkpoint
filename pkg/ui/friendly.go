package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"checkpoint/pkg/disk"
)

var (
	driveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("241")).
			Padding(1).
			Margin(1, 0).
			Width(60)

	driveNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	driveIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	driveDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	progressBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	progressBarFullStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("86")).
				Foreground(lipgloss.Color("16"))

	progressBarEmptyStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("238")).
				Foreground(lipgloss.Color("238"))
)

// DisplayFriendlyDisks shows disks in a Windows-like friendly format
func DisplayFriendlyDisks(groups []disk.DriveGroup) {
	title := titleStyle.Render("💾 My Computer")
	fmt.Println(title)
	fmt.Println()

	for i, group := range groups {
		displayDriveGroup(i+1, group)
	}
}

func displayDriveGroup(id int, group disk.DriveGroup) {
	// Create drive content
	content := ""
	
	// Header with icon and name
	header := fmt.Sprintf("%s %s %s",
		driveIconStyle.Render(group.Icon),
		driveNameStyle.Render(group.Name),
		driveDescStyle.Render(fmt.Sprintf("(%s)", group.Description)))
	content += header + "\n\n"
	
	// Size information
	usedPercent := float64(group.TotalUsed) / float64(group.TotalSize) * 100
	content += fmt.Sprintf("📊 Space: %s free of %s\n",
		availableStyle.Render(FormatBytes(group.Available)),
		sizeStyle.Render(FormatBytes(group.TotalSize)))
	
	// Progress bar
	content += "\n" + createProgressBar(int(usedPercent), 40) + fmt.Sprintf(" %.1f%%", usedPercent) + "\n"
	
	// Mount points
	if len(group.Disks) == 1 {
		content += fmt.Sprintf("\n📁 Location: %s", group.Disks[0].MountPoint)
	} else {
		content += fmt.Sprintf("\n📁 Locations:")
		for _, disk := range group.Disks {
			content += fmt.Sprintf("\n   • %s (%s)", disk.MountPoint, FormatBytes(disk.Size))
		}
	}
	
	// Special badges
	if group.IsPrimary {
		content += "\n\n" + availableStyle.Render("⭐ Primary Drive")
	}
	
	// Apply box style
	box := driveBoxStyle.Render(content)
	fmt.Println(box)
}

func createProgressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	
	filled := width * percent / 100
	empty := width - filled
	
	bar := progressBarFullStyle.Render(strings.Repeat("█", filled)) +
		progressBarEmptyStyle.Render(strings.Repeat("░", empty))
	
	return bar
}

// DisplaySimpleDiskList shows a simplified disk list
func DisplaySimpleDiskList(groups []disk.DriveGroup) {
	headers := []string{"ID", "Name", "Size", "Free", "Used", "Type"}
	headerRow := makeSimpleRow(headers, headerStyle)
	fmt.Println(headerRow)

	for i, group := range groups {
		usedPercent := float64(group.TotalUsed) / float64(group.TotalSize) * 100
		rowData := []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%s %s", group.Icon, group.Name),
			sizeStyle.Render(FormatBytes(group.TotalSize)),
			availableStyle.Render(FormatBytes(group.Available)),
			usedStyle.Render(fmt.Sprintf("%.1f%%", usedPercent)),
			group.Description,
		}
		
		style := rowStyle
		if i%2 == 0 {
			style = evenRowStyle
		}
		fmt.Println(makeSimpleStyledRow(rowData, style))
	}
}