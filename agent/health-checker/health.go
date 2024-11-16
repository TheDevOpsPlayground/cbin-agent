package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

// Config holds the configuration for the health checker
type Config struct {
	RecycleBinDir    string `json:"recycleBinDir"`
	NumWorkers       int    `json:"numWorkers"`
	CheckIntervalSec int    `json:"checkIntervalSec"`
	MaxRetries       int    `json:"maxRetries"`
	RetryDelaySec    int    `json:"retryDelaySec"`
}

// HealthStatus represents the current health status of cbin
type HealthStatus struct {
	Timestamp         time.Time `json:"timestamp"`
	ProgramRunning    bool      `json:"program_running"`
	RecycleBinExists  bool      `json:"recycle_bin_exists"`
	RecycleFileExists bool      `json:"recycle_file_exists"`
	AliasExists       bool      `json:"alias_exists"`
	NFSExists         bool      `json:"nfs_exists"`
	LastError         string    `json:"last_error,omitempty"`
}

type HealthChecker struct {
	config      Config
	status      HealthStatus
	statusMutex sync.RWMutex
	logger      *logrus.Logger
	stopChan    chan struct{}
	httpServer  *http.Server // Add this field
}

const (
	healthCheckPort = ":10001" // Using the original port from health.go
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	// Create log file
	logFile := &lumberjack.Logger{
		Filename:   "/var/log/cbin/health-checker.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	}
	logger.SetOutput(logFile)

	// Load configuration
	config, err := loadConfig("/etc/cbin/config.conf")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize health checker
	checker := NewHealthChecker(config, logger)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start health checker
	go checker.Start()

	// Start HTTP server
	go checker.startHTTPServer()

	// Wait for shutdown signal
	<-sigChan
	checker.Stop()
}

func NewHealthChecker(config Config, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
		status: HealthStatus{
			Timestamp: time.Now(),
		},
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(time.Duration(hc.config.CheckIntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.performHealthCheck()
		case <-hc.stopChan:
			return
		}
	}
}

func (hc *HealthChecker) Stop() {
	// Signal the health check goroutine to stop
	close(hc.stopChan)

	// Shutdown the HTTP server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hc.httpServer.Shutdown(ctx); err != nil {
		hc.logger.Errorf("Error shutting down HTTP server: %v", err)
	}
}

func (hc *HealthChecker) performHealthCheck() {
	hc.statusMutex.Lock()
	defer hc.statusMutex.Unlock()

	hc.status.Timestamp = time.Now()

	// Check if cbin process is running
	hc.status.ProgramRunning = hc.checkProcessRunning()

	// Check if recycle bin directory exists
	hc.status.RecycleBinExists = hc.checkRecycleBinExists()

	// Check if cbin executable exists
	hc.status.RecycleFileExists = hc.checkRecycleFileExists()

	// Check if rm alias exists
	hc.status.AliasExists = hc.checkAliasExists()

	// Check NFS mount
	hc.status.NFSExists = hc.checkNFSWithRetries()

	// If any critical check fails, remove the alias
	if !hc.status.ProgramRunning || !hc.status.NFSExists {
		hc.logger.Warn("Critical health check failed, removing rm alias")
		if err := hc.removeAlias(); err != nil {
			hc.logger.Errorf("Failed to remove alias: %v", err)
		}
	}

	hc.logStatus()
}

func (hc *HealthChecker) checkProcessRunning() bool {
	cmd := exec.Command("pgrep", "-f", "cbin")
	if err := cmd.Run(); err != nil {
		hc.status.LastError = fmt.Sprintf("cbin process not running: %v", err)
		return false
	}
	return true
}

func (hc *HealthChecker) checkRecycleBinExists() bool {
	if _, err := os.Stat(hc.config.RecycleBinDir); err != nil {
		hc.status.LastError = fmt.Sprintf("recycle bin directory not accessible: %v", err)
		return false
	}
	return true
}

func (hc *HealthChecker) checkRecycleFileExists() bool {
	if _, err := os.Stat("/usr/local/bin/cbin"); err != nil {
		hc.status.LastError = fmt.Sprintf("cbin executable not found: %v", err)
		return false
	}
	return true
}

func (hc *HealthChecker) checkNFSWithRetries() bool {
	for i := 0; i < hc.config.MaxRetries; i++ {
		if hc.checkNFSMount() {
			return true
		}
		hc.logger.Warnf("NFS check failed, attempt %d/%d", i+1, hc.config.MaxRetries)
		time.Sleep(time.Duration(hc.config.RetryDelaySec) * time.Second)
	}
	hc.status.LastError = "NFS mount check failed after max retries"
	return false
}

func (hc *HealthChecker) checkNFSMount() bool {
	// Use df command to check if any NFS mounts exist
	out, err := exec.Command("df", "-t", "nfs", "-t", "nfs4").Output()
	if err != nil {
		hc.status.LastError = fmt.Sprintf("failed to check NFS mount: %v", err)
		hc.logger.Errorf("Failed to check NFS mount: %v", err)
		return false
	}
	// If we have any output, it means NFS mounts exist
	return len(out) > 0
}

func (hc *HealthChecker) checkAliasExists() bool {
	content, err := ioutil.ReadFile("/etc/bash.bashrc")
	if err != nil {
		hc.status.LastError = fmt.Sprintf("failed to read bash.bashrc: %v", err)
		hc.logger.Errorf("Failed to read bash.bashrc: %v", err)
		return false
	}

	hc.logger.Debugf("Content of bash.bashrc: %s", string(content))

	aliasLine := `alias rm='/usr/local/bin/cbin'`
	exists := strings.Contains(string(content), aliasLine)

	hc.logger.Debugf("Alias line '%s' exists: %v", aliasLine, exists)

	return exists
}

func (hc *HealthChecker) removeAlias() error {
	content, err := ioutil.ReadFile("/etc/bash.bashrc")
	if err != nil {
		return fmt.Errorf("failed to read bash.bashrc: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, `alias rm='/usr/local/bin/cbin'`) {
			newLines = append(newLines, line)
		}
	}

	err = ioutil.WriteFile("/etc/bash.bashrc", []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write bash.bashrc: %v", err)
	}

	// Execute source command to reload bash configuration
	cmd := exec.Command("bash", "-c", "source /etc/bash.bashrc")
	return cmd.Run()
}

func (hc *HealthChecker) logStatus() {
	hc.logger.WithFields(logrus.Fields{
		"status":          hc.status,
		"program_running": hc.status.ProgramRunning,
		"nfs_exists":      hc.status.NFSExists,
		"alias_exists":    hc.status.AliasExists,
		"timestamp":       hc.status.Timestamp,
	}).Info("Health check completed")
}

func (hc *HealthChecker) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", hc.healthHandler)

	hc.httpServer = &http.Server{
		Addr:         healthCheckPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	hc.logger.Infof("Starting health check server on port %s", healthCheckPort)

	if err := hc.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		hc.logger.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (hc *HealthChecker) healthHandler(w http.ResponseWriter, r *http.Request) {
	hc.statusMutex.RLock()
	defer hc.statusMutex.RUnlock()

	// Set timeout header
	w.Header().Set("Content-Type", "application/json")

	// Check if context is done/canceled
	select {
	case <-r.Context().Done():
		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
		return
	default:
		if err := json.NewEncoder(w).Encode(hc.status); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func loadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %v", err)
	}

	// Set default values if not specified
	if config.CheckIntervalSec == 0 {
		config.CheckIntervalSec = 60 // Default to 1 minute
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 5 // Default to 5 retries
	}
	if config.RetryDelaySec == 0 {
		config.RetryDelaySec = 60 // Default to 1 minute between retries
	}

	return config, nil
}
