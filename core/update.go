package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
}

var RemoteVer string

const (
	githubAPIURL    = "https://api.github.com/repos/demonkingswarn/luffy/releases/latest"
	releasesURL     = "https://github.com/demonkingswarn/luffy/releases/download"
	updateUserAgent = "luffy-updater/1.0"
)

// checkUpdate fetches the latest release version from GitHub
func checkUpdate() error {
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", updateUserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release data: %w", err)
	}

	RemoteVer = strings.TrimPrefix(release.TagName, "v")
	return nil
}

// getBinaryPath finds the path to the current luffy executable
func getBinaryPath() (string, error) {
	exeName := "luffy"
	if runtime.GOOS == "windows" {
		exeName = "luffy.exe"
	}

	// Try to find in PATH first
	bin, err := exec.LookPath(exeName)
	if err == nil && bin != "" {
		return filepath.Abs(bin)
	}

	// Fallback to current executable path
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine executable path: %w", err)
	}

	return filepath.Abs(exe)
}

// getDownloadURL returns the appropriate download URL for the current OS/arch
func getDownloadURL(version string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "linux":
		return fmt.Sprintf("%s/v%s/luffy-linux-%s", releasesURL, version, arch)
	case "windows":
		return fmt.Sprintf("%s/v%s/luffy-windows-%s.exe", releasesURL, version, arch)
	case "darwin":
		return fmt.Sprintf("%s/v%s/luffy-macos-%s", releasesURL, version, arch)
	case "freebsd":
		return fmt.Sprintf("%s/v%s/luffy-freebsd-%s", releasesURL, version, arch)
	case "android":
		return fmt.Sprintf("%s/v%s/luffy-android-%s", releasesURL, version, arch)
	default:
		return ""
	}
}

// downloadUpdate downloads a URL to a local file with validation
func downloadUpdate(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer out.Close()

	// Copy with validation
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(filepath)
		return fmt.Errorf("download incomplete: %w", err)
	}

	if written == 0 {
		os.Remove(filepath)
		return fmt.Errorf("downloaded file is empty")
	}

	return nil
}

// Update checks for and installs the latest version
func Update() error {
	fmt.Println("Checking for updates...")

	if err := checkUpdate(); err != nil {
		return err
	}

	currentVer := strings.TrimPrefix(Version, "v")
	remoteVer := RemoteVer

	if currentVer == remoteVer {
		fmt.Println("Already up to date! (" + Version + ")")
		return nil
	}

	fmt.Printf("Update available: %s -> v%s\n", Version, RemoteVer)

	// Get binary path
	binPath, err := getBinaryPath()
	if err != nil {
		return err
	}

	// Get download URL
	downloadURL := getDownloadURL(remoteVer)
	if downloadURL == "" {
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Create temp file for download
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("luffy-update-%s.tmp", remoteVer))

	fmt.Println("Downloading update...")
	if err := downloadUpdate(downloadURL, tempFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify download
	info, err := os.Stat(tempFile)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("cannot verify download: %w", err)
	}

	if info.Size() < 1024 {
		os.Remove(tempFile)
		return fmt.Errorf("downloaded file is too small, may be corrupt")
	}

	// Set executable permissions on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile, 0755); err != nil {
			os.Remove(tempFile)
			return fmt.Errorf("cannot set permissions: %w", err)
		}
	}

	// Handle Windows special case - cannot overwrite running binary
	if runtime.GOOS == "windows" {
		return updateWindows(binPath, tempFile)
	}

	// Standard update for Unix systems
	return updateUnix(binPath, tempFile)
}

// updateUnix performs the update on Unix-like systems
func updateUnix(binPath, tempFile string) error {
	// Create backup
	backupPath := binPath + ".backup"
	if err := os.Rename(binPath, backupPath); err != nil {
		return fmt.Errorf("cannot create backup: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tempFile, binPath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, binPath)
		return fmt.Errorf("cannot install update: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	fmt.Printf("Successfully updated to v%s!\n", RemoteVer)
	return nil
}

// updateWindows handles the Windows update using a batch script
func updateWindows(binPath, tempFile string) error {
	// Create batch script for delayed replacement
	batchScript := filepath.Join(os.TempDir(), "luffy-update.bat")
	script := fmt.Sprintf(`@echo off
chcp 65001 >nul
timeout /t 1 /nobreak >nul
move /Y "%s" "%s" >nul 2>&1
if errorlevel 1 (
    echo Failed to update. Please update manually.
    pause
) else (
    echo Update complete!
)
del "%%~f0"
`, tempFile, binPath)

	if err := os.WriteFile(batchScript, []byte(script), 0755); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("cannot create update script: %w", err)
	}

	// Execute batch script and exit
	fmt.Println("Update downloaded. Applying changes...")
	cmd := exec.Command("cmd", "/C", "start", "/B", batchScript)
	if err := cmd.Start(); err != nil {
		os.Remove(tempFile)
		os.Remove(batchScript)
		return fmt.Errorf("cannot apply update: %w", err)
	}

	fmt.Println("Update will be applied when you restart luffy.")
	os.Exit(0)
	return nil // Never reached
}
