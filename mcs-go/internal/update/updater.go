package update

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
	"time"

	"github.com/michaelkeevildown/mcs/internal/version"
)

// ReleaseInfo represents a GitHub release
type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckForUpdates checks if a newer version is available
func CheckForUpdates(currentVersion string) {
	// Skip update check for development builds
	if version.IsDevBuild() {
		return
	}

	// Check for updates in the background
	go func() {
		latest, err := GetLatestRelease()
		if err != nil {
			return
		}

		if IsNewerVersion(currentVersion, latest.TagName) {
			// Store update info for later use
			storeUpdateInfo(latest)
		}
	}()
}

// GetLatestRelease fetches the latest release from GitHub
func GetLatestRelease() (*ReleaseInfo, error) {
	url := "https://api.github.com/repos/michaelkeevildown/michaels-codespaces/releases/latest"
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	
	return &release, nil
}

// GetDevRelease fetches the development release
func GetDevRelease() (*ReleaseInfo, error) {
	url := "https://api.github.com/repos/michaelkeevildown/michaels-codespaces/releases/tags/dev-latest"
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	
	return &release, nil
}

// IsNewerVersion compares two version strings
func IsNewerVersion(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	// Simple string comparison for now
	// TODO: Implement proper semantic version comparison
	return current != latest && latest > current
}

// DownloadAndInstall downloads and installs a specific release
func DownloadAndInstall(release *ReleaseInfo) error {
	// Find the appropriate asset for this platform
	assetName := fmt.Sprintf("mcs-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".zip"
	} else {
		assetName += ".tar.gz"
	}
	
	var asset *Asset
	for _, a := range release.Assets {
		if a.Name == assetName {
			asset = &a
			break
		}
	}
	
	if asset == nil {
		return fmt.Errorf("no release found for platform %s-%s", runtime.GOOS, runtime.GOARCH)
	}
	
	// Download to temporary file
	tmpFile, err := os.CreateTemp("", "mcs-update-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Download the asset
	resp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}
	
	// Extract and install
	if err := extractAndInstall(tmpFile.Name(), assetName); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}
	
	return nil
}

// extractAndInstall extracts the downloaded archive and installs the binary
func extractAndInstall(archivePath, archiveName string) error {
	tmpDir, err := os.MkdirTemp("", "mcs-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	
	// Extract archive
	if strings.HasSuffix(archiveName, ".zip") {
		cmd := exec.Command("unzip", "-q", archivePath, "-d", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		cmd := exec.Command("tar", "-xzf", archivePath, "-C", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	}
	
	// Find the binary
	binaryName := "mcs"
	if runtime.GOOS == "windows" {
		binaryName = "mcs.exe"
	}
	
	extractedBinary := filepath.Join(tmpDir, binaryName)
	if _, err := os.Stat(extractedBinary); err != nil {
		return fmt.Errorf("binary not found in archive: %w", err)
	}
	
	// Get install path
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		home, _ := os.UserHomeDir()
		mcsHome = filepath.Join(home, ".mcs")
	}
	
	installPath := filepath.Join(mcsHome, "bin", "mcs")
	
	// Backup current binary
	if _, err := os.Stat(installPath); err == nil {
		backupPath := installPath + ".backup"
		if err := os.Rename(installPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}
		// Remove backup on successful install
		defer func() {
			if _, err := os.Stat(installPath); err == nil {
				os.Remove(backupPath)
			} else {
				// Restore backup on failure
				os.Rename(backupPath, installPath)
			}
		}()
	}
	
	// Install new binary
	if err := copyFile(extractedBinary, installPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}
	
	// Make executable
	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}
	
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	
	_, err = io.Copy(destination, source)
	return err
}

// storeUpdateInfo stores update information for later use
func storeUpdateInfo(release *ReleaseInfo) {
	// TODO: Implement storing update info for displaying to user
	// This could be stored in a file or displayed immediately
}

// BuildFromSource updates MCS by building from source
func BuildFromSource() error {
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		home, _ := os.UserHomeDir()
		mcsHome = filepath.Join(home, ".mcs")
	}
	
	// Pull latest changes
	cmd := exec.Command("git", "pull", "origin", "main")
	cmd.Dir = mcsHome
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, output)
	}
	
	// Build
	mcsGoDir := filepath.Join(mcsHome, "mcs-go")
	cmd = exec.Command("make", "install")
	cmd.Dir = mcsGoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %w\nOutput: %s", err, output)
	}
	
	return nil
}