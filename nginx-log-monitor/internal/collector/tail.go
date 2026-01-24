package collector

import (
	"fmt"
	"nginx-log-monitor/internal/model"
	"time"

	"github.com/nxadm/tail"
)

// LogCollector handles the tailing of log files.
type LogCollector struct {
	filePath string
	tail     *tail.Tail
	// DataChannel is used to send raw log lines to the parser.
	DataChannel chan string
	stopChan    chan struct{}
}

// NewLogCollector creates a new instance of LogCollector.
func NewLogCollector(path string) (*LogCollector, error) {
	// Config for tailing
	// ReOpen: true to follow log rotation
	// Follow: true to keep waiting for new lines
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: false, // File might not exist yet
		Poll:      true,  // Poll for changes (sometimes needed for Docker/mounted volumes)
	}

	t, err := tail.TailFile(path, config)
	if err != nil {
		return nil, fmt.Errorf("failed to tail file: %v", err)
	}

	return &LogCollector{
		filePath:    path,
		tail:        t,
		DataChannel: make(chan string, 1000), // Buffer size can be adjusted
		stopChan:    make(chan struct{}),
	}, nil
}

// Start begins reading lines from the tailed file.
func (lc *LogCollector) Start() {
	go func() {
		defer close(lc.DataChannel)
		for line := range lc.tail.Lines {
			select {
			case <-lc.stopChan:
				return
			default:
				if line.Err != nil {
					fmt.Printf("Error reading line: %v\n", line.Err)
					continue
				}
				// Send the raw text to the channel
				lc.DataChannel <- line.Text
			}
		}
	}()
}

// Stop stops the collector and cleans up resources.
func (lc *LogCollector) Stop() {
	close(lc.stopChan)
	lc.tail.Stop()
}
