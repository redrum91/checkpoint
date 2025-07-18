package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"checkpoint/pkg/disk"
)

var (
	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	summaryTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")).
				MarginBottom(1)

	summaryItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	summaryValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
)

func DisplaySummary(stats disk.DiskStats, disks []disk.Disk) {
	content := summaryTitleStyle.Render("📊 Storage Summary") + "\n\n"

	// Basic stats
	content += formatSummaryLine("Total Disks", fmt.Sprintf("%d", stats.TotalDisks))
	content += formatSummaryLine("Total Capacity", FormatBytes(stats.TotalSize))
	content += formatSummaryLine("Used Space", fmt.Sprintf("%s (%.1f%%)", 
		FormatBytes(stats.TotalUsed), 
		float64(stats.TotalUsed)/float64(stats.TotalSize)*100))
	content += formatSummaryLine("Available", FormatBytes(stats.TotalAvailable))

	// Disk types breakdown
	if len(stats.DisksByType) > 0 {
		content += "\n" + summaryItemStyle.Render("Disk Types:") + "\n"
		for diskType, count := range stats.DisksByType {
			if diskType != disk.TypeSymlink { // Skip symlinks in summary
				icon := getTypeIcon(diskType)
				content += fmt.Sprintf("  %s %s: %s\n", 
					icon, 
					summaryItemStyle.Render(string(diskType)),
					summaryValueStyle.Render(fmt.Sprintf("%d", count)))
			}
		}
	}

	// Main disks summary
	content += "\n" + summaryItemStyle.Render("Main Storage:") + "\n"
	diskCount := 0
	for _, d := range disks {
		if d.Type == disk.TypePhysical || d.Type == disk.TypeLVM {
			diskCount++
			if diskCount <= 5 { // Show first 5 main disks
				name := truncatePath(d.Path, 20)
				usage := float64(d.Used) / float64(d.Size) * 100
				content += fmt.Sprintf("  %s: %s (%s free, %.0f%% used)\n",
					name,
					FormatBytes(d.Size),
					FormatBytes(d.Available),
					usage)
			}
		}
	}
	if diskCount > 5 {
		content += fmt.Sprintf("  ... and %d more\n", diskCount-5)
	}

	// Links summary (condensed)
	if len(stats.Hardlinks) > 0 || len(stats.Symlinks) > 0 {
		content += "\n" + summaryItemStyle.Render("Links:") + "\n"
		if len(stats.Hardlinks) > 0 {
			content += fmt.Sprintf("  🔗 Hard link groups: %s\n", 
				summaryValueStyle.Render(fmt.Sprintf("%d", len(stats.Hardlinks))))
		}
		if len(stats.Symlinks) > 0 {
			content += fmt.Sprintf("  ✨ Symbolic links: %s\n", 
				summaryValueStyle.Render(fmt.Sprintf("%d", len(stats.Symlinks))))
		}
	}

	box := summaryBoxStyle.Render(content)
	fmt.Println(box)
}


func formatSummaryLine(label, value string) string {
	return fmt.Sprintf("%s: %s\n",
		summaryItemStyle.Render(label),
		summaryValueStyle.Render(value))
}

func getTypeIcon(t disk.DiskType) string {
	icons := map[disk.DiskType]string{
		disk.TypePhysical: "💽",
		disk.TypeLVM:      "🗄️",
		disk.TypeLoop:     "🔄",
		disk.TypeBind:     "📁",
		disk.TypeNetwork:  "🌐",
		disk.TypeFUSE:     "🔌",
		disk.TypePath:     "📂",
		disk.TypeManual:   "✋",
	}

	icon, ok := icons[t]
	if !ok {
		return "❓"
	}
	return icon
}