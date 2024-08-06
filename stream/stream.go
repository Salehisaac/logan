package stream

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
	"strings"
	"io"
	"bufio"
	"path/filepath"

)

type LogEntry struct {
    Timestamp time.Time 
    Caller    string 
    Content   string 
    Level     string 
    Trace     string 
    FileName  string
}

var authKey string
var traces []string
var tracesMap = make(map[string][]LogEntry)
var mu sync.Mutex
var client_system string
var pwd string

var base_log_path string

func StartStream(phoneNumber, system, pathToLogs  string ){
	
	client_system = system
	pwd = pathToLogs
    files := []string{"session/access.log", "bff/access.log"}
    resultChan := make(chan string)
    streamTraces(files, resultChan)

    
    clientDirPath := fmt.Sprintf("./stream/logs/%s", phoneNumber) 
    _, err := os.Stat(clientDirPath)
    if os.IsNotExist(err) {
        err = os.MkdirAll(clientDirPath, os.ModePerm)
        if err != nil {
            log.Fatalf("Failed to create directory: %v", err)
        }
    }
   
    nowTime := time.Now().Format("2006-01-02_15-04-05")
    dirName := fmt.Sprintf("./stream/logs/%s/%s_%s",phoneNumber,nowTime, client_system)
    base_log_path = dirName
    err = os.MkdirAll(dirName, os.ModePerm)
    if err != nil {
        log.Fatalf("Failed to create directory: %v", err)
    }

    directories := getLogDirs()
    err = createDirectories(directories, dirName)
    if err != nil {
        log.Fatalf("error creating directories %s", err)
    }
    logFiles := getLogFiles(directories)


    for _, file := range logFiles{
       go streamLogs(file)
    }

	for trace := range resultChan {
		if !traceExists(trace){
            traces = append(traces, trace)
        }
	}
}

func streamTraces(fileNames []string, resultChan chan<-string) {

	var wg sync.WaitGroup
	

	for _, fileName := range fileNames {
		go processFileForTrace(fileName, resultChan, &wg)
	}

}

func streamLogs(fileName string){
   
    parts := strings.Split(fileName, "/")
    dir := parts[len(parts)-2]
    logFile := parts[len(parts)-1]

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
        time.Sleep(1 * time.Second)
		line, err := reader.ReadString('\n')
		if err != nil {
			continue 
		}

		log_trace := extractTraceFromLog(line)


        for _, trace := range traces{
            if log_trace == trace {
                level := extractLevelFromLog(line)
                caller := extractCallerFromLog(line)
                content := extractContentFromLog(line)
                timeStamp := extractTimeFromLog(line)

                log_entry := LogEntry{
                    Timestamp: timeStamp,
                    Content: content,
                    Caller: caller,
                    Trace: trace,
                    Level: level,
                    FileName: fileName,
                }
                fmt.Printf("%s => Caller : %s, Level: %s, Trace: %s\n\n", fileName,caller , level, log_trace)
                pathToWrite := fmt.Sprintf("%s/%s/%s", base_log_path, dir, logFile)
                writeLogs(log_entry, pathToWrite)
                break
            }
        } 
		time.Sleep(100 * time.Millisecond)
	}
}

func processFileForTrace(fileName string, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	filePath := filepath.Join(pwd, fileName)
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
	re := regexp.MustCompile(`perm_auth_key_id:\s*(\d+)`)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			continue 
		}

		trace := extractTraceFromLog(line)
		content := extractContentFromLog(line)
		match := re.FindStringSubmatch(content)

		if len(match) > 0 {
			if !traceExists(trace) {
				resultChan <- trace
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func traceExists(trace string) bool {
    for _, t := range traces {
        if t == trace {
            return true
        }
    }
    return false
}

func getEnteries(line string) (string) {
    line_part := strings.Split(line, " ")
  
    if len(line_part) != 1 {
        log.Fatal("not enough arguments")
    }

    phoneNumber := line_part[0]

    return phoneNumber
}

func checkPathValidation(path string)(bool, error){
    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return false , fmt.Errorf("directory does not exist")
    }
    if err != nil {
        return false , fmt.Errorf("error checking directory: %v", err)
    }
    if !info.IsDir() {
        return false, fmt.Errorf("directory does not exists")
    }
    return true, nil
}

func extractTimeFromLog(line string) (time.Time) {
    var timestamp time.Time
   
    re := regexp.MustCompile(`"@timestamp":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        timestampStr := match[1]
        timestamp, _ = time.Parse(time.RFC3339, timestampStr)
    }

    return timestamp
}
func extractCallerFromLog(line string) (string) {
   
    var caller string

    re := regexp.MustCompile(`"caller":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        caller = match[1]
    }

    return caller
}
func extractLevelFromLog(line string) (string) {
    var  level string

    re := regexp.MustCompile(`"level":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        level = match[1]
    }
   

    return level
}
func getLogDirs() []string {
    directories := []string{"session","status","authsession", "bff", "msg" ,"idgen", "biz", "sync", "media"}

    return directories
}
func getLogFiles(directories []string)[]string{
    var logFiles []string

	
	basePath := pwd

	
	for _, dir := range directories {
		accessLogPath := filepath.Join(basePath, dir, "access.log")
		errorLogPath := filepath.Join(basePath, dir, "error.log")
		logFiles = append(logFiles, accessLogPath, errorLogPath)
	}

	return logFiles
}
func extractContentFromLog(line string) (string) {

    var content string
    re := regexp.MustCompile(`"content":"(.*?)"(?:,|$)`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        content = match[1]
    }

    return content
}
func extractTraceFromLog(line string) (string) {

    var trace string
   
    re := regexp.MustCompile(`"trace":"([a-fA-F0-9]{32})"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        trace = match[1]
    }


    return trace
}
func createDirectories(dirs[]string, base string)error{
    fileNames := []string{"access.log", "error.log"}
    for _, dir := range dirs{
        clientDirPath := fmt.Sprintf("%s/%s" ,base, dir) 
        _, err := os.Stat(clientDirPath)
        if os.IsNotExist(err) {
            err = os.MkdirAll(clientDirPath, os.ModePerm)
            if err != nil {
                return err
            }
        }
        
        for _, fileName := range fileNames{
            filePath := fmt.Sprintf("%s/%s", clientDirPath, fileName)
            file, err := os.Create(filePath)
            if err != nil {
                log.Fatalf("Failed to create file: %v", err)
            }
            defer file.Close()
        }
  
    }
    return nil
}

func writeLogs(log_entry LogEntry, fileName string) {
    file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Failed to open file %s: %v", fileName, err)
    }
    defer file.Close()
   
    logLine := fmt.Sprintf("Timestamp:%s, Content:%s, Caller:%s, Level:%s, Trace:%s\n",
        log_entry.Timestamp, log_entry.Content, log_entry.Caller, log_entry.Level, log_entry.Trace)
    
   
    if _, err := file.WriteString(logLine); err != nil {
        log.Fatalf("Failed to write to file %s: %v", fileName, err)
    }
}