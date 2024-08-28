package stream

import (
	"fmt"
	"io"
	"log"
	"log_reader/internal/types"
	"os"
	"path/filepath"
	"sync"
	"time"

)

type FileInfo struct {
	Time     string
	HasError bool
}

var files = make(map[string]FileInfo)
var mu sync.Mutex



func WriteLogs(log_entry types.LogEntry, fileName string) {
	var clientDirPath string
	var errorFilePath string

	mu.Lock()
	fileInfo, exist := files[log_entry.Trace]
	if !exist {
		tehranLoc, err := time.LoadLocation("Asia/Tehran")
		if err != nil {
			log.Fatal(err)
		}

		tehranTime := log_entry.Timestamp.In(tehranLoc)
		formattedDate := log_entry.Timestamp.Format("2006-01-02")
		formattedTime := tehranTime.Format("15:04:05")
		finalString := fmt.Sprintf("%s_%s", formattedDate, formattedTime)

		clientDirPath = filepath.Join(fileName, fmt.Sprintf("%s_%s.txt", finalString, log_entry.Trace))

		file, err := os.OpenFile(clientDirPath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to create or open file: %v", err)
		}
		file.Close()

		files[log_entry.Trace] = FileInfo{
			Time:     finalString,
			HasError: false,
		}
	} else {
		clientDirPath = filepath.Join(fileName, fmt.Sprintf("%s_%s.txt", fileInfo.Time, log_entry.Trace))
	}
	mu.Unlock()

	if !files[log_entry.Trace].HasError && log_entry.Level == "error" {

		if fileInfo, exists := files[log_entry.Trace]; exists {
			fileInfo.HasError = true
			files[log_entry.Trace] = fileInfo
		}else {
			log.Printf("FileInfo for trace %s does not exist", log_entry.Trace)
		}

		errorFilePath = fmt.Sprintf("%s_error.txt", clientDirPath[:len(clientDirPath)-4])
		err_file, err := os.OpenFile(errorFilePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to create or open file: %v", err)
		}
		source_file, err := os.Open(clientDirPath)
		if err != nil {
			log.Fatalf("Failed to create or open file: %v", err)
		}
		if _, err := io.Copy(err_file, source_file); err != nil {
			log.Fatalf("failed to copy data: %v", err)
		}

		if err := os.Remove(clientDirPath); err != nil {
			log.Fatalf("failed to remove source file: %v", err)
		}
		err_file.Close()
		source_file.Close()
	}

	if files[log_entry.Trace].HasError {
		errorFilePath = filepath.Join(fileName, fmt.Sprintf("%s_%s_error.txt", files[log_entry.Trace].Time, log_entry.Trace))
		clientDirPath = errorFilePath
	}

	file, err := os.OpenFile(clientDirPath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	logLine := fmt.Sprintf("%s => Timestamp:%s, Content:%s, Caller:%s, Level:%s, Trace:%s\n",
		log_entry.FileName, log_entry.Timestamp, log_entry.Content, log_entry.Caller, log_entry.Level, log_entry.Trace)

	if _, err := file.WriteString(logLine); err != nil {
		log.Fatalf("Failed to write to file %s: %v", clientDirPath, err)
	}
}


