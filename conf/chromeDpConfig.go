package conf

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
)

var (
	// Global ChromeDP allocator context
	AllocatorCtx    context.Context
	AllocatorCancel context.CancelFunc
)

// ChromeDPConfig holds Chrome configuration
type ChromeDPConfig struct {
	Headless             bool
	NoSandbox            bool
	DisableGPU           bool
	DisableDevShmUsage   bool
	DisableSetuidSandbox bool
	WindowSize           string
	UserAgent            string
	RemoteDebuggingPort  string
	ChromeBinary         string
}

// GetChromeDPConfig returns optimized Chrome configuration
func GetChromeDPConfig() *ChromeDPConfig {
	config := &ChromeDPConfig{
		Headless:             true,
		NoSandbox:            false,
		DisableGPU:           false,
		DisableDevShmUsage:   true,
		DisableSetuidSandbox: false,
		WindowSize:           "1920,1080",
		UserAgent:            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		RemoteDebuggingPort:  "9222",
		ChromeBinary:         "",
	}

	// Detect if running in Docker or constrained environment
	if isDockerEnvironment() || isConstrainedEnvironment() {
		config.NoSandbox = true
		config.DisableGPU = true
		config.DisableDevShmUsage = true
		config.DisableSetuidSandbox = true
	}

	// Override with environment variables if set
	if headless := os.Getenv("CHROMEDP_HEADLESS"); headless != "" {
		config.Headless = strings.ToLower(headless) == "true"
	}

	if noSandbox := os.Getenv("CHROMEDP_NO_SANDBOX"); noSandbox != "" {
		config.NoSandbox = strings.ToLower(noSandbox) == "true"
	}

	if disableGPU := os.Getenv("CHROMEDP_DISABLE_GPU"); disableGPU != "" {
		config.DisableGPU = strings.ToLower(disableGPU) == "true"
	}

	if userAgent := os.Getenv("CHROMEDP_USER_AGENT"); userAgent != "" {
		config.UserAgent = userAgent
	}

	if chromeBin := os.Getenv("CHROME_BIN"); chromeBin != "" {
		config.ChromeBinary = chromeBin
	}

	if windowSize := os.Getenv("CHROMEDP_WINDOW_SIZE"); windowSize != "" {
		config.WindowSize = windowSize
	}

	return config
}

// isDockerEnvironment detects if running inside Docker
func isDockerEnvironment() bool {
	// Check for Docker environment indicators
	indicators := []string{
		"DOCKER_ENV",
		"KUBERNETES_SERVICE_HOST",
		"CONTAINER",
	}

	for _, indicator := range indicators {
		if os.Getenv(indicator) != "" {
			return true
		}
	}

	// Check for /.dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") || strings.Contains(content, "containerd") {
			return true
		}
	}

	return false
}

// isConstrainedEnvironment detects resource-constrained environments
func isConstrainedEnvironment() bool {
	// Check available memory (less than 1GB indicates constrained environment)
	if memInfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(memInfo), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if memKB, err := strconv.Atoi(fields[1]); err == nil {
						memMB := memKB / 1024
						if memMB < 1024 { // Less than 1GB
							return true
						}
					}
				}
				break
			}
		}
	}

	// Check CPU count (single core might indicate constrained environment)
	if runtime.NumCPU() == 1 {
		return true
	}

	return false
}

// BuildChromeDPOptions builds ChromeDP allocator options from config
func (c *ChromeDPConfig) BuildChromeDPOptions() []chromedp.ExecAllocatorOption {
	opts := []chromedp.ExecAllocatorOption{
		// Basic flags
		chromedp.Flag("headless", c.Headless),
		chromedp.Flag("disable-gpu", c.DisableGPU),
		chromedp.Flag("disable-dev-shm-usage", c.DisableDevShmUsage),

		// Performance optimizations
		chromedp.Flag("no-sandbox", c.NoSandbox),
		chromedp.Flag("disable-setuid-sandbox", c.DisableSetuidSandbox),
		chromedp.Flag("single-process", true), // Mode single process untuk headless
		chromedp.Flag("no-zygote", true),      // Nonaktifkan zygote process
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-background-networking", true),

		// Network optimizations
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),

		// Rendering optimizations
		chromedp.Flag("disable-features", "TranslateUI,BlinkGenPropertyTrees,site-per-process"),
		chromedp.Flag("blink-settings", "imagesEnabled=false"), // Nonaktifkan gambar
		chromedp.Flag("force-color-profile", "srgb"),

		// Window/User Agent
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent(c.UserAgent),
	}

	// Chrome binary path
	if c.ChromeBinary != "" {
		opts = append(opts, chromedp.ExecPath(c.ChromeBinary))
	}

	// Remote debugging
	if c.RemoteDebuggingPort != "" {
		opts = append(opts,
			chromedp.Flag("remote-debugging-port", c.RemoteDebuggingPort),
			chromedp.Flag("remote-debugging-address", "0.0.0.0"),
		)
	}

	return opts
}

// InitializeChromeDPAllocator initializes the global ChromeDP allocator
func InitializeChromeDPAllocator() error {
	config := GetChromeDPConfig()
	opts := config.BuildChromeDPOptions()

	// Create allocator context
	AllocatorCtx, AllocatorCancel = chromedp.NewExecAllocator(context.Background(), opts...)

	return nil
}

// GetChromeDPContext creates a new ChromeDP context using the global allocator
func GetChromeDPContext() (context.Context, context.CancelFunc) {
	if AllocatorCtx == nil {
		// Fallback initialization if not called in init()
		err := InitializeChromeDPAllocator()
		if err != nil {
			fmt.Println("Error initializing ChromeDP allocator:", err)
			return nil, nil
		}
	}

	return chromedp.NewContext(AllocatorCtx)
}

// Cleanup closes the global allocator
func Cleanup() {
	if AllocatorCancel != nil {
		AllocatorCancel()
	}
}
