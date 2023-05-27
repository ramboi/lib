package lib

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc/peer"
)

type Config struct {
	data map[string]interface{}
	sync.RWMutex
}

// Definig max file size for logfile.txt to 1MB
const (
	maxLogFileSize = 1 * 1024 * 1024
	directory_path = "/opt/bazc/bazcli/log"
)

func (c *Config) Init() {
	c.data = make(map[string]interface{})
}

func Log_Init(logFileName string) error {
	filePath := filepath.Join(directory_path, logFileName)
	file, err := OpenLogFile(filePath)
	if err != nil {
		return err
	}
	log.SetOutput(file)
	defer file.Close()
	return err
}

func OpenLogFile(logFileName string) (*os.File, error) {
	file, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() >= maxLogFileSize {
		err := file.Close()
		if err != nil {
			return nil, err
		}
		rotateLogFile(logFileName)

		file, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}

func rotateLogFile(logFileName string) error {
	dir, fileName := filepath.Split(logFileName)
	ext := filepath.Ext(fileName)
	baseName := fileName[:len(fileName)-len(ext)]
	timestamp := time.Now().Format("20060102150405")
	newFileName := filepath.Join(dir, baseName+"_"+timestamp+ext)

	// Rename the current log file to the new file name
	err := os.Rename(logFileName, newFileName)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Put(ctx context.Context, value interface{}) {
	if p, ok := peer.FromContext(ctx); ok {
		c.Lock()
		defer c.Unlock()
		c.data[p.Addr.String()] = value
		return
	}
	panic("unable to get peer from context.")
}

func (c *Config) Get(ctx context.Context) interface{} {
	if p, ok := peer.FromContext(ctx); ok {
		c.RLock()
		defer c.RUnlock()
		return c.data[p.Addr.String()]
	}
	panic("unable to get peer from context.")
}

func (c *Config) Delete(ctx context.Context) {
	if p, ok := peer.FromContext(ctx); ok {
		c.Lock()
		defer c.Unlock()
		delete(c.data, p.Addr.String())
		return
	}
	panic("unable to get peer from context.")
}
