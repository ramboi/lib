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

var Directory_path = "/opt/bazc/bazcli/log"

// Definig max file size for logfile.txt to 10KB
const (
	maxLogFileSize = 1 * 1024 * 1024
)

func (c *Config) Init() {
	c.data = make(map[string]interface{})
}

func Log_Init(logFileName string) (*os.File, error) {
	full_path := filepath.Join(Directory_path, logFileName)
	file, err := OpenLogFile(full_path)
	if err != nil {
		return nil, err
	}
	log.SetOutput(file)
	log.Println("Logging start.")

}
func OpenLogFile(logFileName string) (*os.File, error) {
	file, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
		return rotateLogFile(logFileName)
	}
	return file, nil
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

func rotateLogFile(logFileName string) (*os.File, error) {
	dir, fileName := filepath.Split(logFileName)
	ext := filepath.Ext(fileName)
	baseName := fileName[:len(fileName)-len(ext)]
	timestamp := time.Now().Format("20060102150405")
	newFileName := filepath.Join(dir, baseName+"_"+timestamp+ext)

	// Rename the current log file to the new file name
	err := os.Rename(logFileName, newFileName)
	if err != nil {
		return nil, err
	}

	// Create a new log file
	file, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}
