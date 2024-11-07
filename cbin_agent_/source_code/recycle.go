package main

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "time"
)

var (
    recycleBin = "/mnt/test/" 
    logFile    = "/mnt/test/recycle.log" 
)


func init() {
    if _, err := os.Stat(recycleBin); os.IsNotExist(err) {
        os.MkdirAll(recycleBin, 0777) 
    }

    logDir := filepath.Dir(logFile)
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        os.MkdirAll(logDir, 0755) 
    }
}

func logAction(action string, filePath string) {
    f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        fmt.Printf("Error opening log file: %v\n", err)
        return
    }
    defer f.Close()

    logLine := fmt.Sprintf("%s - %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), action, filePath)
    if _, err := f.WriteString(logLine); err != nil {
        fmt.Printf("Error writing to log file: %v\n", err)
    } else {
        fmt.Println("Log saved to", logFile)
    }
}

func moveToRecycleBin(filePath string, recursive bool, force bool) {
    if _, err := os.Stat(filePath); err == nil {
        newPath := filepath.Join(recycleBin, filepath.Base(filePath))
        err := os.Rename(filePath, newPath)
        if err != nil {
            fmt.Printf("Error moving file: %v\n", err)
            return
        }
        logAction("Moved", filePath)
        fmt.Printf("Moved '%s' to recycle bin.\n", filePath)
    } else {
        fmt.Printf("File '%s' not found.\n", filePath)
    }
}

func main() {
    recursive := flag.Bool("r", false, "Remove directories recursively")
    force := flag.Bool("f", false, "Force remove files or directories")
    flag.Parse()

    files := flag.Args()

    if len(files) == 0 {
        fmt.Println("Usage: recycle [-r] [-f] <file1> <file2> ...")
        os.Exit(1)
    }

    for _, file := range files {
        moveToRecycleBin(file, *recursive, *force)
    }
}
