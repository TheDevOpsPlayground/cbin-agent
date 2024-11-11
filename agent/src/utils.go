package main

import (
	"fmt"
	"net"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func initLogger() {
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/cbin/cbin.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
	})
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Set file permission to 0777
	err := os.Chmod("/var/log/cbin/cbin.log", 0777)
	if err != nil {
		logrus.Errorf("Failed to set permissions on /var/log/cbin/cbin.log: %v", err)
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
