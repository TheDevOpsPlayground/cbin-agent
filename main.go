package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
)

type Config struct {
	RecycleBinDir string `json:"recycleBinDir"`
	NumWorkers    int    `json:"numWorkers"`
}

func main() {
	// Define flags
	var (
		help  = flag.Bool("h", false, "Show help")
		files = flag.String("f", "", "Comma-separated list of files to recycle (e.g., file1.txt,file2.log)")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Check if files flag is provided
	if *files == "" {
		log.Println("Please specify files to recycle using -f.")
		printHelp()
		return
	}

	// Read configuration from file
	config, err := readConfig("/etc/recycler-cli/config.conf")
	if err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	// Set up logging with log rotation
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/recycler-cli.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	// Get server's private IP and hostname
	ip, hostname, err := getServerInfo()
	if err != nil {
		log.Fatalf("Failed to get server information: %v", err)
	}

	// Create unique directory for the server
	serverDir := filepath.Join(config.RecycleBinDir, fmt.Sprintf("%s_%s", ip, hostname))
	err = os.MkdirAll(serverDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create server directory: %v", err)
	}

	// Split comma-separated files into a slice
	fileSlice := strings.Split(*files, ",")

	// Create a channel to send file paths to workers
	fileChan := make(chan string, len(fileSlice))
	for _, file := range fileSlice {
		fileChan <- strings.TrimSpace(file)
	}
	close(fileChan)

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				moveFileToRecycleBin(file, serverDir)
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	log.Println("All specified files have been recycled.")
}

func moveFileToRecycleBin(file string, serverDir string) {
	// Check if file exists
	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Skipping non-existent file: %s\n", file)
		} else {
			log.Printf("Error checking existence of %s: %v\n", file, err)
		}
		return
	}

	// Get current date and time
	now := time.Now()
	dateDir := filepath.Join(serverDir, now.Format("2006-01-02"))
	err = os.MkdirAll(dateDir, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create date directory: %v", err)
		return
	}

	// Construct destination path with timestamp
	timestamp := now.Format("15:04:05")
	destPath := filepath.Join(dateDir, fmt.Sprintf("%s_%s%s", fileInfo.Name(), timestamp, filepath.Ext(file)))

	// Move file to recycle bin
	log.Printf("Moving %s to %s...\n", file, destPath)
	err = os.Rename(file, destPath)
	if err != nil {
		log.Printf("Failed to move %s to recycle bin: %v\n", file, err)
		log.Println("Operation incomplete. Check files and try again.")
		return
	}
	log.Printf("%s successfully moved to recycle bin.\n", file)
}

func readConfig(filePath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getServerInfo() (string, string, error) {
	// Get server's private IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if !ip.IsLoopback() && ip.To4() != nil {
				return ip.String(), "", nil
			}
		}
	}
	return "", "", fmt.Errorf("no private IP found")
}

func printHelp() {
	fmt.Println("recycler-cli - A safer alternative to rm, moving files to a recycle bin.")
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("Usage:")
	fmt.Println("  recycler-cli [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -f, --files   Comma-separated list of files to recycle (e.g., file1.txt,file2.log)")
	fmt.Println("  -h, --help    Display this help message")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  recycler-cli -f file1.txt,file2.log,file3.pdf")
	fmt.Println()
	fmt.Println("Important:")
	fmt.Println("  - Ensure the recycle bin directory is set to a valid path.")
	fmt.Println("  - Be cautious with file paths to avoid unintended actions.")
}
