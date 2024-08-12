package logreader

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log_reader/configs"
	"log_reader/internal/types"
	"log_reader/pkg/utils"
	stream_utils "log_reader/pkg/utils/stream"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	mu     sync.Mutex
	traces []string
	logs   = make(map[string][]types.LogEntry)
)

func StartStream(cfg *configs.Config) {
	files := []string{"session/access.log", "bff/access.log"}
	resultChan := make(chan string)

	clientDirPath := fmt.Sprintf("./stream/logs/%s", cfg.PhoneNumber)
	_, err := os.Stat(clientDirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(clientDirPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	nowTime := time.Now().Format("2006-01-02_15-04-05")
	dirName := fmt.Sprintf("./stream/logs/%s/%s_%s", cfg.PhoneNumber, nowTime, cfg.ClientSystem)
	err = os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	directories := utils.GetLogDirs()
	// err = utils.CreateDirectories(directories, dirName)
	// if err != nil {
	//     log.Fatalf("error creating directories %s", err)
	// }
	logFiles := utils.GetLogFiles(directories, cfg)

	StreamTraces(files, resultChan, cfg)

	for _, file := range logFiles {
		go func(file string) {
			streamLogs(file)
		}(file)
	}

	for trace := range resultChan {
		if !utils.TraceExists(trace, traces) {
			traces = append(traces, trace)
			time.Sleep(1 * time.Second)
			go func(trace string) {
				checkLogsWithTraces(trace, dirName)
			}(trace)
		}
	}
}

func StreamTraces(fileNames []string, resultChan chan<- string, cfg *configs.Config) {

	for _, fileName := range fileNames {
		go processFileForTrace(fileName, resultChan, cfg)
	}
}

func processFileForTrace(fileName string, resultChan chan<- string, cfg *configs.Config) {

	filePath := filepath.Join(cfg.LogsPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("Error seeking to end of file %s: %v\n", filePath, err)
		return
	}

	reader := bufio.NewReader(file)
	re := regexp.MustCompile(`perm_auth_key_id:\s*(-?\d+)`)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trace := utils.ExtractTraceFromLog(line)
		content := utils.ExtractContentFromLog(line)
		match := re.FindStringSubmatch(content)

		if len(match) > 0 && match[1] == cfg.AuthKey {
			if !utils.TraceExists(trace, traces) {
				resultChan <- trace
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func streamLogs(fileName string) {
	parts := strings.Split(fileName, "/")
	dir := parts[len(parts)-2]
	logFile := parts[len(parts)-1]
	dirLogFile := fmt.Sprintf("%s/%s", dir, logFile)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fileName, err)
		return
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("Error seeking to end of file %s: %v\n", fileName, err)
		return
	}

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Printf("Error reading line from file %s: %v\n", fileName, err)
			return
		}

		log_trace := utils.ExtractTraceFromLog(line)
		level := utils.ExtractLevelFromLog(line)
		caller := utils.ExtractCallerFromLog(line)
		content := utils.ExtractContentFromLog(line)
		timeStamp := utils.ExtractTimeFromLog(line)

		log_entry := types.LogEntry{
			Timestamp: timeStamp,
			Content:   content,
			Caller:    caller,
			Trace:     log_trace,
			Level:     level,
			FileName:  fileName,
		}
		mu.Lock()
		logs[dirLogFile] = append(logs[dirLogFile], log_entry)
		mu.Unlock()
	}
}

func checkLogsWithTraces(trace, dirname string) {
	for fileName, fileLogs := range logs {
		for i := len(fileLogs) - 1; i >= 0; i-- {
			log := fileLogs[i]
			if log.Trace == trace {
				fmt.Printf("%s => Caller: %s, Level: %s, Trace: %s\n\n", fileName, log.Caller, log.Level, log.Trace)
				pathToWrite := fmt.Sprintf("%s", dirname)
				stream_utils.WriteLogs(log, pathToWrite)
				mu.Lock()
				fileLogs = append(fileLogs[:i], fileLogs[i+1:]...)
				mu.Unlock()
			}
		}
	}
}
