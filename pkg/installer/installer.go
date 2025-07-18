package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"checkpoint/pkg/disk"
)

// ExecuteCommand runs an installation command with optional target disk
func ExecuteCommand(command string, targetDisk *disk.Disk) error {
	// Security check: warn about sudo
	if strings.Contains(command, "sudo") {
		fmt.Printf("\n⚠️  Warning: Command contains 'sudo'. Consider running without elevated privileges.\n")
		fmt.Printf("Press Enter to continue or Ctrl+C to cancel...")
		fmt.Scanln()
	}

	fmt.Printf("\n🚀 Executing: %s\n", command)
	if targetDisk != nil {
		fmt.Printf("📍 Target location: %s\n", targetDisk.MountPoint)
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create command without automatic sudo handling
	cmd := exec.Command(parts[0], parts[1:]...)

	// Set working directory and environment if target disk specified
	if targetDisk != nil {
		// Check write permissions
		if !canWrite(targetDisk.MountPoint) {
			return fmt.Errorf("no write permission for %s", targetDisk.MountPoint)
		}

		cmd.Dir = targetDisk.MountPoint
		
		// Add common installation environment variables
		env := os.Environ()
		env = append(env,
			fmt.Sprintf("PREFIX=%s/.local", os.Getenv("HOME")),
			fmt.Sprintf("DESTDIR=%s", targetDisk.MountPoint),
			fmt.Sprintf("INSTALL_ROOT=%s", targetDisk.MountPoint),
		)
		cmd.Env = env
	}

	// Connect to standard streams
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	
	// Check if the error is due to permissions
	if err != nil && isPermissionError(command, err) {
		return promptForSudo(command, targetDisk)
	}
	
	return err
}

// isPermissionError checks if the error is likely due to missing permissions
func isPermissionError(command string, err error) bool {
	// Check for common permission-related exit codes and commands
	if exitErr, ok := err.(*exec.ExitError); ok {
		// dpkg/apt returns 100 for permission errors
		if exitErr.ExitCode() == 100 {
			return true
		}
		// Generic permission denied is often exit code 1 or 2
		if (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 2) &&
			(strings.Contains(command, "apt") || strings.Contains(command, "dpkg") ||
			 strings.Contains(command, "yum") || strings.Contains(command, "dnf") ||
			 strings.Contains(command, "pacman") || strings.Contains(command, "zypper")) {
			return true
		}
	}
	return false
}

// promptForSudo asks the user if they want to retry with sudo
func promptForSudo(command string, targetDisk *disk.Disk) error {
	fmt.Printf("\n⚠️  The command failed, likely due to missing permissions.\n")
	fmt.Printf("\n🔒 SECURITY WARNING:\n")
	fmt.Printf("   Running commands with sudo gives them full system access.\n")
	fmt.Printf("   Only proceed if you trust this command and understand the risks.\n")
	fmt.Printf("\nWould you like to retry with sudo? (yes/no): ")
	
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response == "yes" || response == "y" {
		// Prepend sudo if not already present
		if !strings.HasPrefix(command, "sudo ") {
			command = "sudo " + command
		}
		
		fmt.Printf("\n🚀 Executing with sudo: %s\n", command)
		
		parts := strings.Fields(command)
		cmd := exec.Command(parts[0], parts[1:]...)
		
		// Note: When using sudo, we can't set custom environment variables
		// or working directory as effectively
		if targetDisk != nil {
			fmt.Printf("⚠️  Note: Target disk settings may not apply with sudo\n")
		}
		
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		
		return cmd.Run()
	}
	
	return fmt.Errorf("command cancelled by user")
}

// canWrite checks if the current user can write to a directory
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

// DetectPackageManager detects the system's package manager
func DetectPackageManager() string {
	managers := []struct {
		name string
		cmd  string
	}{
		{"apt", "apt"},
		{"dnf", "dnf"},
		{"yum", "yum"},
		{"zypper", "zypper"},
		{"pacman", "pacman"},
		{"apk", "apk"},
		{"emerge", "emerge"},
	}

	for _, pm := range managers {
		if _, err := exec.LookPath(pm.cmd); err == nil {
			return pm.name
		}
	}

	return "unknown"
}

// SuggestInstallCommand suggests an installation command based on the package manager
func SuggestInstallCommand(packageName string) string {
	pm := DetectPackageManager()
	
	commands := map[string]string{
		"apt":    fmt.Sprintf("sudo apt install %s", packageName),
		"dnf":    fmt.Sprintf("sudo dnf install %s", packageName),
		"yum":    fmt.Sprintf("sudo yum install %s", packageName),
		"zypper": fmt.Sprintf("sudo zypper install %s", packageName),
		"pacman": fmt.Sprintf("sudo pacman -S %s", packageName),
		"apk":    fmt.Sprintf("sudo apk add %s", packageName),
		"emerge": fmt.Sprintf("sudo emerge %s", packageName),
	}

	if cmd, ok := commands[pm]; ok {
		return cmd
	}

	return fmt.Sprintf("# Package manager not detected. Manual installation required for %s", packageName)
}