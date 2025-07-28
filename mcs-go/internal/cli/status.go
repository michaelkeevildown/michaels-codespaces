package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

var (
	statusHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	statusLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	dividerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	runningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	stoppedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// StatusCommand creates the 'status' command
func StatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "status [name]",
		Aliases: []string{"monitor"},
		Short:   "📊 Show system and codespace status",
		Long:    "Display comprehensive system information and codespace status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return showCodespaceStatus(args[0])
			}
			return showSystemStatus()
		},
	}
}

func showSystemStatus() error {
	// Don't clear screen - let user scroll back if needed
	
	// Show beautiful header
	ui.ShowHeader()
	
	// System Information
	fmt.Println(statusHeaderStyle.Render("📊 SYSTEM INFORMATION"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 21)))
	
	// Hostname
	hostname, _ := os.Hostname()
	fmt.Printf("🖥️  Hostname: %s\n", hostname)
	
	// Architecture
	fmt.Printf("🏗️  Architecture: %s\n", runtime.GOARCH)
	
	// OS
	hostInfo, _ := host.Info()
	osInfo := getOSInfo(hostInfo)
	fmt.Printf("🐧 OS: %s\n", osInfo)
	
	// Uptime
	uptime := time.Duration(hostInfo.Uptime) * time.Second
	fmt.Printf("⏰ Uptime: %s\n", formatUptime(uptime))
	fmt.Println()
	
	// System Resources
	fmt.Println(statusHeaderStyle.Render("💻 SYSTEM RESOURCES"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 21)))
	
	// Memory
	vmStat, _ := mem.VirtualMemory()
	memUsage := fmt.Sprintf("%s/%s (%.1f%%)", 
		formatBytesStatus(int64(vmStat.Used)), 
		formatBytesStatus(int64(vmStat.Total)),
		vmStat.UsedPercent)
	fmt.Printf("💾 Memory: %s\n", memUsage)
	
	// CPU
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		fmt.Printf("🔥 CPU Load: %.1f%%\n", cpuPercent[0])
	}
	
	// Disk
	diskStat, _ := disk.Usage("/")
	diskUsage := fmt.Sprintf("%s/%s (%s used)", 
		formatBytesStatus(int64(diskStat.Used)), 
		formatBytesStatus(int64(diskStat.Total)),
		fmt.Sprintf("%.0f%%", diskStat.UsedPercent))
	fmt.Printf("💿 Disk: %s\n", diskUsage)
	fmt.Println()
	
	// Docker Status
	fmt.Println(statusHeaderStyle.Render("🐳 DOCKER STATUS"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 21)))
	
	// Check Docker
	if err := showDockerStatus(); err != nil {
		fmt.Println(errorStyle.Render("❌ Docker: Not running"))
	}
	fmt.Println()
	
	// Codespaces
	fmt.Println(statusHeaderStyle.Render("🚀 CODESPACES"))
	fmt.Println(dividerStyle.Render(strings.Repeat("─", 21)))
	
	ctx := context.Background()
	showCodespacesOverview(ctx)
	
	fmt.Println()
	fmt.Println(strings.Repeat("═", 63))
	fmt.Println()
	fmt.Println("Quick commands:")
	fmt.Println("  • mcs list          - List all codespaces")
	fmt.Println("  • mcs start <name>  - Start a codespace")
	fmt.Println("  • mcs stop <name>   - Stop a codespace")
	fmt.Println("  • mcs create <name> - Create new codespace")
	fmt.Println()
	
	return nil
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d minutes", minutes))
	}
	
	return strings.Join(parts, ", ")
}

func getOSInfo(hostInfo *host.InfoStat) string {
	if runtime.GOOS == "darwin" {
		// Try to get macOS version
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			return fmt.Sprintf("macOS %s", strings.TrimSpace(string(out)))
		}
		return fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
	}
	
	// For Linux, try lsb_release first
	out, err := exec.Command("lsb_release", "-d").Output()
	if err == nil {
		parts := strings.Split(string(out), ":")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	
	return fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
}

func showDockerStatus() error {
	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	
	ctx := context.Background()
	info, err := client.GetSystemInfo(ctx)
	if err != nil {
		return err
	}
	
	fmt.Println(runningStyle.Render("✅ Docker: Running"))
	
	// Get version from docker command
	cmd := exec.Command("docker", "--version")
	if output, err := cmd.Output(); err == nil {
		version := strings.TrimSpace(string(output))
		// Extract just the version number
		if parts := strings.Split(version, " "); len(parts) >= 3 {
			fmt.Printf("📦 Version: %s\n", strings.TrimSuffix(parts[2], ","))
		}
	}
	
	// Count running containers
	containers, err := client.ListContainers(ctx, "")
	runningCount := 0
	if err == nil {
		for _, c := range containers {
			if c.State == "running" {
				runningCount++
			}
		}
	}
	
	fmt.Printf("🏃 Containers: %d running, %d total\n", runningCount, info.Containers)
	fmt.Printf("🖼️  Images: %d\n", info.Images)
	
	return nil
}

func showCodespacesOverview(ctx context.Context) {
	manager := codespace.NewManager()
	codespaces, err := manager.List(ctx)
	if err != nil {
		fmt.Println(errorStyle.Render("⚠️  Unable to list codespaces"))
		return
	}
	
	if len(codespaces) == 0 {
		fmt.Println(dimStyle.Render("No codespaces directory found"))
		return
	}
	
	// Count running codespaces
	runningCount := 0
	for _, cs := range codespaces {
		if cs.Status == "running" {
			runningCount++
		}
	}
	
	fmt.Printf("📁 Total Codespaces: %d\n", len(codespaces))
	fmt.Printf("✅ Running: %d\n", runningCount)
	
	if len(codespaces) > 0 {
		fmt.Println()
		fmt.Println(sectionStyle.Render("📋 CODESPACE LIST"))
		fmt.Println(dividerStyle.Render(strings.Repeat("─", 21)))
		
		for _, cs := range codespaces {
			fmt.Println()
			fmt.Printf("📦 %s\n", cs.Name)
			
			if cs.Status == "running" {
				fmt.Printf("   Status: %s\n", runningStyle.Render("🟢 Running"))
				fmt.Printf("   VS Code: %s\n", urlStyle.Render(cs.VSCodeURL))
			} else {
				fmt.Printf("   Status: %s\n", stoppedStyle.Render("⭕ Stopped"))
				fmt.Printf("   URL: Port %d\n", cs.VSCodePort)
			}
			
			if cs.Repository != "" {
				fmt.Printf("   Repo: %s\n", cs.Repository)
			}
		}
	}
}

func showCodespaceStatus(name string) error {
	ctx := context.Background()
	manager := codespace.NewManager()
	
	cs, err := manager.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("codespace '%s' not found", name)
	}
	
	fmt.Println()
	fmt.Println(statusHeaderStyle.Render("Codespace Status"))
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()
	
	fmt.Printf("📦 Name: %s\n", cs.Name)
	
	if cs.Status == "running" {
		fmt.Printf("📊 Status: %s\n", runningStyle.Render("🟢 Running"))
		fmt.Printf("🔗 VS Code: %s\n", urlStyle.Render(cs.VSCodeURL))
		fmt.Printf("🌐 App URL: %s\n", urlStyle.Render(cs.AppURL))
		
		// Get container stats
		client, err := docker.NewClient()
		if err == nil {
			defer client.Close()
			
			containers, err := client.ListContainers(ctx, fmt.Sprintf("mcs.name=%s", cs.Name))
			if err == nil && len(containers) > 0 {
				container := containers[0]
				
				fmt.Println()
				fmt.Println(sectionStyle.Render("Container Details:"))
				fmt.Printf("   ID: %s\n", container.ID[:12])
				fmt.Printf("   Image: %s\n", container.Image)
				fmt.Printf("   Created: %s\n", container.Created)
				
				// Show resource usage if available
				if stats := container.Stats; stats != nil {
					fmt.Printf("   CPU: %.1f%%\n", stats.CPUPercent)
					fmt.Printf("   Memory: %s / %s (%.1f%%)\n", 
						formatBytesStatus(int64(stats.MemoryUsage)), 
						formatBytesStatus(int64(stats.MemoryLimit)),
						stats.MemoryPercent)
				}
			}
		}
	} else {
		fmt.Printf("📊 Status: %s\n", stoppedStyle.Render("⭕ Stopped"))
		fmt.Printf("🔗 VS Code Port: %d\n", cs.VSCodePort)
	}
	
	if cs.Repository != "" {
		fmt.Printf("📁 Repository: %s\n", cs.Repository)
	}
	
	fmt.Printf("📍 Location: %s\n", cs.Path)
	
	// Show directory size
	cmd := exec.Command("du", "-sh", cs.Path+"/src")
	if output, err := cmd.Output(); err == nil {
		parts := strings.Fields(string(output))
		if len(parts) > 0 {
			fmt.Printf("💾 Source size: %s\n", parts[0])
		}
	}
	
	fmt.Println()
	fmt.Println(dimStyle.Render("Tip: Use 'mcs recover " + cs.Name + "' for quick credential recovery"))
	fmt.Println()
	
	return nil
}

func formatBytesStatus(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}