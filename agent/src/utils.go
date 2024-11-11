package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type MultiWriter struct {
	writers []io.Writer
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

func initLogger(serverDir string) {
	var writers []io.Writer

	// Primary log file
	primaryLog := &lumberjack.Logger{
		Filename:   "/var/log/cbin/cbin.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
	}
	writers = append(writers, primaryLog)

	// Secondary log file
	secondaryLog := &lumberjack.Logger{
		Filename:   filepath.Join(serverDir, "cbin.log"),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
	}
	writers = append(writers, secondaryLog)

	multiWriter := &MultiWriter{writers: writers}
	logrus.SetOutput(multiWriter)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Set file permission to 0777 for both log files
	err := os.Chmod("/var/log/cbin/cbin.log", 0777)
	if err != nil {
		logrus.Errorf("Failed to set permissions on /var/log/cbin/cbin.log: %v", err)
	}
	err = os.Chmod(filepath.Join(serverDir, "cbin.log"), 0777)
	if err != nil {
		logrus.Errorf("Failed to set permissions on %s: %v", filepath.Join(serverDir, "cbin.log"), err)
	}
}

func getServerInfo() (string, string, error) {
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
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() && v.IP.To4() != nil {
					hostname, err := os.Hostname()
					if err != nil {
						return "", "", err
					}
					return v.IP.String(), hostname, nil
				}
			}
		}
	}
	return "", "", fmt.Errorf("no private IP found")
}
