package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// DockerImageInfo contains information about a Docker image
type DockerImageInfo struct {
	Repository   string
	Tag          string
	ImageID      string
	Created      time.Time
	Size         string
	LocalVersion string
}

// CodeServerRelease represents a GitHub release
type CodeServerRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
}

// UpdateChecker handles checking for Docker image updates
type UpdateChecker struct {
	httpClient *http.Client
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestCodeServerVersion fetches the latest code-server release from GitHub
func (u *UpdateChecker) GetLatestCodeServerVersion() (*CodeServerRelease, error) {
	resp, err := u.httpClient.Get("https://api.github.com/repos/coder/code-server/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release CodeServerRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release info: %w", err)
	}

	return &release, nil
}

// GetLocalCodeServerVersion gets the version of the local code-server image
func (u *UpdateChecker) GetLocalCodeServerVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "json", "codercom/code-server:latest")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get docker image info: %w", err)
	}

	if len(output) == 0 {
		return "", fmt.Errorf("codercom/code-server:latest image not found")
	}

	// Parse the JSON output
	var imageInfo struct {
		ID        string `json:"ID"`
		CreatedAt string `json:"CreatedAt"`
	}
	if err := json.Unmarshal(output, &imageInfo); err != nil {
		return "", fmt.Errorf("failed to parse docker image info: %w", err)
	}

	// Try to get the version from image labels
	cmd = exec.CommandContext(ctx, "docker", "inspect", "codercom/code-server:latest", "--format", "{{index .Config.Labels \"org.opencontainers.image.version\"}}")
	versionOutput, err := cmd.Output()
	if err == nil && len(versionOutput) > 0 {
		version := strings.TrimSpace(string(versionOutput))
		if version != "" && version != "<no value>" {
			return version, nil
		}
	}

	// Fallback: try to get version from running the image
	cmd = exec.CommandContext(ctx, "docker", "run", "--rm", "codercom/code-server:latest", "--version")
	versionOutput, err = cmd.CombinedOutput()
	if err == nil {
		// Parse version from output like "4.102.2 abcd1234 with Code 1.102.2"
		// The output may have some log lines before the version
		lines := strings.Split(string(versionOutput), "\n")
		for _, line := range lines {
			// Look for line containing version number pattern
			versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)\s+[a-f0-9]+\s+with`)
			matches := versionRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				return "v" + matches[1], nil
			}
		}
	}

	return "", fmt.Errorf("unable to determine local code-server version")
}

// CheckForUpdates checks if there are updates available for code-server
func (u *UpdateChecker) CheckForUpdates(ctx context.Context) (bool, *CodeServerRelease, string, error) {
	latest, err := u.GetLatestCodeServerVersion()
	if err != nil {
		return false, nil, "", fmt.Errorf("failed to get latest version: %w", err)
	}

	local, err := u.GetLocalCodeServerVersion(ctx)
	if err != nil {
		// If we can't determine local version, assume update is available
		return true, latest, "unknown", nil
	}

	// Compare versions
	if local != latest.TagName {
		return true, latest, local, nil
	}

	return false, latest, local, nil
}

// GetMCSImages returns information about all MCS Docker images
func (u *UpdateChecker) GetMCSImages(ctx context.Context) ([]DockerImageInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "json", "--filter", "reference=mcs/code-server-*")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list MCS images: %w", err)
	}

	var images []DockerImageInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}

		var imgData struct {
			ID         string `json:"ID"`
			Repository string `json:"Repository"`
			Tag        string `json:"Tag"`
			CreatedAt  string `json:"CreatedAt"`
			Size       string `json:"Size"`
		}

		if err := json.Unmarshal([]byte(line), &imgData); err != nil {
			continue
		}

		created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", imgData.CreatedAt)
		
		images = append(images, DockerImageInfo{
			Repository: imgData.Repository,
			Tag:        imgData.Tag,
			ImageID:    imgData.ID,
			Created:    created,
			Size:       imgData.Size,
		})
	}

	return images, nil
}

// PullLatestCodeServer pulls the latest code-server image
func (u *UpdateChecker) PullLatestCodeServer(ctx context.Context, progress func(string)) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", "codercom/code-server:latest")
	
	if progress != nil {
		progress("Pulling latest code-server image...")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image: %w\nOutput: %s", err, string(output))
	}

	if progress != nil {
		progress("Successfully pulled latest code-server image")
	}

	return nil
}

// RebuildMCSImages rebuilds all MCS Docker images
func (u *UpdateChecker) RebuildMCSImages(ctx context.Context, dockerfilesPath string, progress func(string)) error {
	buildScript := dockerfilesPath + "/build.sh"
	
	cmd := exec.CommandContext(ctx, "bash", buildScript)
	cmd.Dir = dockerfilesPath

	if progress != nil {
		progress("Rebuilding MCS Docker images...")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rebuild images: %w\nOutput: %s", err, string(output))
	}

	if progress != nil {
		progress("Successfully rebuilt all MCS images")
	}

	return nil
}

// CompareVersions compares two version strings (e.g., "v4.102.2" and "v4.103.0")
func CompareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return len(parts1) - len(parts2)
}