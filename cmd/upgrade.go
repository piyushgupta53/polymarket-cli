package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/piyushgupta/polymarket-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	upgradeCheckOnly bool
	upgradeForce     bool
)

const githubRepo = "piyushgupta/polymarket-cli"

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for updates and upgrade the CLI",
	Long:  "Check GitHub Releases for a newer version of polymarket-cli and optionally download it.",
	RunE:  runUpgrade,
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeCheckOnly, "check", false, "Check for updates without installing")
	upgradeCmd.Flags().BoolVar(&upgradeForce, "force", false, "Skip confirmation prompt")
	rootCmd.AddCommand(upgradeCmd)
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Body    string        `json:"body"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	current := appVersion

	log.Debug("checking for updates", "current", current)

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")

	if getOutputFormat() == "json" {
		return output.PrintJSON(map[string]string{
			"current_version": current,
			"latest_version":  latest,
			"up_to_date":      fmt.Sprintf("%t", current == latest),
		})
	}

	if current == latest || current == "dev" && !upgradeForce {
		if current == "dev" {
			fmt.Println("Running development build. Use --force to upgrade anyway.")
			fmt.Printf("Latest release: %s\n", latest)
			return nil
		}
		fmt.Printf("Already up to date (%s)\n", current)
		return nil
	}

	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Latest version:  %s\n", latest)

	if release.Body != "" {
		fmt.Printf("\nChangelog:\n%s\n", release.Body)
	}

	if upgradeCheckOnly {
		return nil
	}

	// Find matching asset
	assetName := expectedAssetName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s (expected %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	if !upgradeForce {
		if IsNonInteractive() {
			return output.ErrInvalidInput("upgrade requires --force in non-interactive mode")
		}
		fmt.Printf("\nDownload %s? This will replace the current binary. [y/N] ", assetName)
		var answer string
		_, _ = fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Download and replace
	fmt.Printf("Downloading %s...\n", assetName)
	if err := downloadAndReplace(downloadURL); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	fmt.Printf("Upgraded to %s\n", latest)
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release: %w", err)
	}

	return &release, nil
}

func expectedAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	ext := ""
	if os == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("polymarket_%s_%s%s", os, arch, ext)
}

func downloadAndReplace(url string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed (status %d)", resp.StatusCode)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current binary: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	// Write to temp file in same directory
	dir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(dir, "polymarket-upgrade-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing binary: %w", err)
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("backing up current binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	// Remove backup
	_ = os.Remove(backupPath)

	return nil
}
