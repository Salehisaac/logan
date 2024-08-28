package static

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log_reader/internal/types"
	"log_reader/pkg/utils"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tracesMap = make(map[string][]types.LogEntry)
	mu        sync.Mutex
)

func GetTheDuration(since string) (time.Duration, error) {

	var duration time.Duration

	timeParts := strings.Split(since, ":")
	if len(timeParts) != 3 {
		return duration, fmt.Errorf("invalid time pattern")
	}

	hours, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return duration, fmt.Errorf("invalid hour input: %v", err)
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return duration, fmt.Errorf("invalid minute input: %v", err)
	} else if minutes > 60 {
		return duration, fmt.Errorf("invalid minute input: %v", err)
	}

	seconds, err := strconv.Atoi(timeParts[2])
	if err != nil {
		return duration, fmt.Errorf("invalid second input: %v", err)
	} else if seconds > 60 {
		return duration, fmt.Errorf("invalid second input: %v", err)
	}

	duration = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

	return duration, nil
}

func Read(offset int64, limit int64, fileName string, pastTime time.Time, traces []string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Seek(offset, 0)
	reader := bufio.NewReader(file)

	var cumulativeSize int64

	for {
		if cumulativeSize >= limit {
			break
		}

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		cumulativeSize += int64(len(line))
		line = strings.TrimSpace(line)
		if line != "" {

			timestamp := utils.ExtractTimeFromLog(line)

			if timestamp.After(pastTime) && timestamp.Before(time.Now()) {

				content := utils.ExtractContentFromLog(line)
				caller := utils.ExtractCallerFromLog(line)
				trace := utils.ExtractTraceFromLog(line)
				level := utils.ExtractLevelFromLog(line)

				log_entry := types.LogEntry{
					Timestamp: timestamp,
					Content:   content,
					Caller:    caller,
					Trace:     trace,
					Level:     level,
					FileName:  fileName,
				}

				for _, t := range traces {
					if t == trace {
						mu.Lock()
						tracesMap[trace] = append(tracesMap[trace], log_entry)
						mu.Unlock()
					}
				}
			}
		}
	}
}
func WriteLogsToFiles(dirName string) {
	tehranLoc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		log.Fatal(err)
	}
	for trace, logs := range tracesMap {

		var hasError bool

		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Timestamp.Before(logs[j].Timestamp)
		})

		for _, log := range logs {
			if log.Level == "error" {
				hasError = true
				break
			}
		}

		earliestTimestamp := logs[0].Timestamp
		formattedDate := earliestTimestamp.Format("2006-01-02")

		var formattedTime string

		if earliestTimestamp.Location() == tehranLoc {
			formattedTime = earliestTimestamp.Format("15:04:05")
		} else {
			tehranTime := earliestTimestamp.In(tehranLoc)
			formattedTime = tehranTime.Format("15:04:05")
		}
		var fileName string
		if hasError {
			fileName = fmt.Sprintf("%s/%s_%s_%s_error.txt", dirName, formattedDate, formattedTime, trace)
		} else {
			fileName = fmt.Sprintf("%s/%s_%s_%s.txt", dirName, formattedDate, formattedTime, trace)
		}
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer file.Close()

		for _, logEntry := range logs {
			logLine := fmt.Sprintf("%s => Timestamp:%s , Content:%s , Caller:%s , Level:%s , Trace:%s", logEntry.FileName, logEntry.Timestamp, logEntry.Content, logEntry.Caller, logEntry.Level, logEntry.Trace)
			file.WriteString(string(logLine) + "\n")
		}
	}
	fmt.Println()
	fmt.Println("Files have been written in logs dir")
}
