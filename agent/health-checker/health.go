package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	clientIP      string
	serverPort    = ":10001"
	retryCount    = 5
	retryDelay    = 1 * time.Minute
	recycleBinDir string
)

type HealthResponse struct {
	ProgramRunning      bool   `json:"program_running"`
	RecycleBinExists    bool   `json:"recycle_bin_exists"`
	RecycleFileExists   bool   `json:"recycle_file_exists"`
	AliasExists         bool   `json:"alias_exists"`
	NFSExists           bool   `json:"nfs_exists"`
	OverallHealthStatus string `json:"overall_health_status"`
}

type Config struct {
	RecycleBinDir string `json:"recycleBinDir"`
	NumWorkers    int    `json:"numWorkers"`
}

func main() {
	envFile := "/etc/cbin/env"
	loadEnv(envFile)

	configFile := "/etc/cbin/config.conf"
	loadConfig(configFile)

	startHTTPServer()
}

func loadEnv(envFile string) {
	file, err := os.Open(envFile)
	if err != nil {
		fmt.Println("Error opening env file:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "client_ip":
			clientIP = value
		}
	}
}

func loadConfig(configFile string) {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		os.Exit(1)
	}

	recycleBinDir = config.RecycleBinDir
}

func startHTTPServer() {
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Starting HTTP server on", clientIP+serverPort)
	err := http.ListenAndServe(clientIP+serverPort, nil)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	healthResponse := evaluateHealth()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse)
}

// evaluateHealth dynamically checks each condition and updates the status accordingly.
func evaluateHealth() HealthResponse {
	healthResponse := HealthResponse{
		ProgramRunning:    true,
		RecycleBinExists:  checkRecycleBin(),
		RecycleFileExists: checkRecycleFile(),
		AliasExists:       checkAlias(),
		NFSExists:         checkNFSWithRetries(),
	}

	// If any condition fails, set the overall health status to "failed"
	if !healthResponse.ProgramRunning || !healthResponse.RecycleBinExists || !healthResponse.RecycleFileExists ||
		!healthResponse.AliasExists || !healthResponse.NFSExists {
		healthResponse.OverallHealthStatus = "failed"
	} else {
		healthResponse.OverallHealthStatus = "ok"
	}

	return healthResponse
}

func checkRecycleBin() bool {
	if recycleBinDir == "" {
		fmt.Println("Recycle bin directory not set in config file.")
		return false
	}

	if _, err := os.Stat(recycleBinDir); err == nil {
		return true
	}
	return false
}

func checkRecycleFile() bool {
	if _, err := os.Stat("/opt/cbin/cbin"); err == nil {
		return true
	}
	return false
}

func checkAlias() bool {
	file, err := os.Open("/etc/bash.bashrc")
	if err != nil {
		fmt.Println("Error opening bash.bashrc:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "alias rm='/opt/cbin/cbin'") {
			return true
		}
	}
	return false
}

// checkNFSWithRetries checks NFS and retries if it fails
func checkNFSWithRetries() bool {
	for i := 0; i < retryCount; i++ {
		if checkNFS() {
			return true
		}
		fmt.Printf("NFS check failed. Retrying in %v... (Attempt %d/%d)\n", retryDelay, i+1, retryCount)
		time.Sleep(retryDelay)
	}

	// If all retries fail, remove alias and reload bashrc
	removeAliasAndReload()
	return false
}

func checkNFS() bool {
	out, err := exec.Command("df", "-h").Output()
	if err != nil {
		fmt.Println("Error running df command:", err)
		return false
	}

	expectedMount := clientIP + ":/mnt/check/" + clientIP
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, expectedMount) {
			return true
		}
	}
	return false
}

func removeAliasAndReload() {
	input, err := ioutil.ReadFile("/etc/bash.bashrc")
	if err != nil {
		fmt.Println("Error reading bash.bashrc:", err)
		return
	}

	output := ""
	for _, line := range strings.Split(string(input), "\n") {
		if !strings.Contains(line, "alias rm='/etc/c_bin/recycle'") {
			output += line + "\n"
		}
	}

	err = ioutil.WriteFile("/etc/bash.bashrc", []byte(output), 0644)
	if err != nil {
		fmt.Println("Error writing to bash.bashrc:", err)
		return
	}

	exec.Command("bash", "-c", "source /etc/bash.bashrc").Run()
}
