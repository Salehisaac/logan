package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
    "io"
	"github.com/joho/godotenv"
    "log_reader/stream"
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
var streamFlag bool


const mb = 1024 * 1024

func init() {
    flag.BoolVar(&streamFlag, "s", false, "Activate stream mode")
}
func main(){
    flag.Parse()
    err := godotenv.Load() 
    if err != nil {
        log.Println("Error loading .env file")
    }

	path := os.Getenv("LOGS_PATH")
	fmt.Print("phone/duration: ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := strings.TrimSpace(scanner.Text())
    phoneNumber , since := getEnteries(line)
    phoneNumber = strings.TrimSpace(phoneNumber)

    if strings.HasPrefix(phoneNumber, "09") {
        phoneNumber = "98" + phoneNumber[1:]
    }

    duration, err := getTheDuration(since)
    if err != nil {
        log.Fatal("error parsing the time : " , err)
    }
 
    if ok, err :=checkPathValidation(path); !ok {
        log.Fatal(" path invalid: " , err)
    }else{
        pwd = path
    }
   
    b, err := NewBot()
    if err != nil {
        log.Fatalf("Failed to create bot: %v", err)
    }

    userId, err := b.GetUserByPhoneNumber(phoneNumber)
    if err != nil {
        log.Fatalf("Failed to get user ID: %v", err)
    }


   
	authKeys, err := b.GetAuthkeysByUserId(userId)
	if err != nil {
        log.Fatalf("Failed to get user ID: %v", err)
    }

    sortedAuthKeys , err := b.SortAuthkeys(authKeys)
  
    if err != nil {
        log.Fatal("failed to sort the authkeys : ", err)
    }

   
 
	clients , err := b.GetDevicesByAuthKeyId(sortedAuthKeys)
	if err != nil {
		log.Fatal("error getting the client : " ,err)
	}


	fmt.Println("choose a op system :")


	for index , client :=range clients{
		fmt.Printf("%d)%s_%s_%s\n",index+1 ,client.DeviceModel, client.SystemVersion, client.ClientIp)
	}
	var opIndex int 
	fmt.Scanln(&opIndex)

	for index , client := range clients{
		if index+1 == opIndex {
			authKeyint := client.AuthKeyId
			authKey = strconv.Itoa(authKeyint)
            client_system = strings.Replace(client.SystemVersion, " ", "_", -1)
		}
	}
	if authKey == ""{
		fmt.Println("invalid number")
		return
	}

    if streamFlag{
        stream.StartStream(phoneNumber, client_system, pwd)
    }else{

        start := time.Now()

        currentTime := time.Now()
        pastTime := currentTime.Add(-duration)
    
    
        clientDirPath := fmt.Sprintf("./logs/%s", phoneNumber) 
        _, err = os.Stat(clientDirPath)
        if os.IsNotExist(err) {
            err = os.MkdirAll(clientDirPath, os.ModePerm)
            if err != nil {
                log.Fatalf("Failed to create directory: %v", err)
            }
        }
       
        pastTimeStr := pastTime.Format("2006-01-02_15-04-05")
        dirName := fmt.Sprintf("./logs/%s/%s_logs_%s",phoneNumber, pastTimeStr, client_system)
        err = os.MkdirAll(dirName, os.ModePerm)
        if err != nil {
            log.Fatalf("Failed to create directory: %v", err)
        }
    
    
    
        directories := getLogDirs()
        files := getLogFiles(directories)
    
        fmt.Println("Processing...")
        processTraces(pastTime) 
    
        var wg sync.WaitGroup
    
       
        for _, file := range files {
            wg.Add(1) 
            go func(file string) {
                defer wg.Done() 
                processFileLogs(pastTime, file)
            }(file)
        }
        wg.Wait() 
        end := time.Since(start)
    
        writeLogsToFiles(dirName)
        
        fmt.Printf("Application runtime: %v\n", end)

    }
   
}
func processTraces(pastTime time.Time) {

    
    dir := filepath.Join(pwd, "session/access.log")
    dir2 := filepath.Join(pwd, "bff/access.log")
    rootDires := []string{dir, dir2}

    
    for _, filePath := range rootDires{
        var wg sync.WaitGroup
       
        var limit int64 = 2 * mb
    
        file, err := os.Open(filePath)
        if err != nil {
            panic(err)
        }
        defer file.Close()
    
        fileInfo, err := file.Stat()
        if err != nil {
            panic(err)
        }
        fileSize := fileInfo.Size()
    
        numChunks := int(fileSize / limit)
        if fileSize%limit > 0 {
            numChunks++
        }
    
        for i := 0; i < numChunks; i++ {
            wg.Add(1)
            go func(start int64) {
                defer wg.Done()
                readTraces(start, limit, filePath, pastTime)
            }(int64(i) * limit)
        }
    
        wg.Wait()
     }
}
func readTraces(offset int64, limit int64, fileName string, pastTime time.Time){

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

            timestamp:= extractTimeFromLog(line)

            if timestamp.After(pastTime) && timestamp.Before(time.Now()) {

                trace := extractTraceFromLog(line)
                content := extractContentFromLog(line)

                re := regexp.MustCompile(`perm_auth_key_id:\s*(\d+)`)
                match := re.FindStringSubmatch(content)
    
                if len(match) > 0 {
                    mu.Lock()
                    if !traceExists(trace) {
                        traces = append(traces, trace)
                    }
                    mu.Unlock()
                }
            }
		}
	}
}
func processFileLogs(pastTime time.Time, filePath string) {
    var wg sync.WaitGroup

    var limit int64 = 2 * mb

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
    defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

    fileSize := fileInfo.Size()

	numChunks := int(fileSize / limit)
	if fileSize%limit > 0 {
		numChunks++
	}

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(start int64) {
			defer wg.Done()
			read(start, limit, filePath, pastTime)
		}(int64(i) * limit)
	}

	wg.Wait()
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
func traceExists(trace string) bool {
    for _, t := range traces {
        if t == trace {
            return true
        }
    }
    return false
}
func writeLogsToFiles(dirName string) {
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
        tehranTime := earliestTimestamp.In(tehranLoc)
        formattedDate := earliestTimestamp.Format("2006-01-02")
        formattedTime := tehranTime .Format("15:04:05")
        var fileName string
        if hasError {
            fileName = fmt.Sprintf("%s/%s_%s_%s_error.txt", dirName, formattedDate, formattedTime, trace)
        }else{
            fileName = fmt.Sprintf("%s/%s_%s_%s.txt", dirName, formattedDate, formattedTime, trace)
        }
        file, err := os.Create(fileName)
        if err != nil {
            log.Fatalf("Failed to create file: %v", err)
        }
        defer file.Close()

        for _, logEntry := range logs {
            logLine:= fmt.Sprintf("%s => Timestamp:%s , Content:%s , Caller:%s , Level:%s , Trace:%s", logEntry.FileName, logEntry.Timestamp, logEntry.Content, logEntry.Caller, logEntry.Level, logEntry.Trace)
            file.WriteString(string(logLine) + "\n")
        }
    }
    fmt.Println()
    fmt.Println("Files have been written in logs dir")
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
func read(offset int64, limit int64, fileName string, pastTime time.Time) {
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

            timestamp:= extractTimeFromLog(line)

            if timestamp.After(pastTime) && timestamp.Before(time.Now()) {

                content := extractContentFromLog(line)
                caller := extractCallerFromLog(line)
                trace := extractTraceFromLog(line)
                level := extractLevelFromLog(line)

                log_entry := LogEntry{
                    Timestamp: timestamp,
                    Content: content,
                    Caller: caller,
                    Trace: trace,
                    Level: level,
                    FileName: fileName,
                }
              
                for _ , t :=range traces{
                    if t == trace {
                        mu.Lock()
                        tracesMap[trace] = append(tracesMap[trace],log_entry)
                        mu.Unlock()
                    }
                }
            }
		}
	}
}
func getEnteries(line string) (string, string) {
    line_part := strings.Split(line, " ")
  
    if len(line_part) != 2 {
        log.Fatal("not enough arguments")
    }

    phoneNumber := line_part[0]
    duration := line_part[1]
    

    return phoneNumber, duration
}
func getTheDuration(since string)(time.Duration, error){

    var duration time.Duration

    timeParts := strings.Split(since, ":")
	if len(timeParts) != 3 {
		return duration , fmt.Errorf("invalid time pattern")
	}

	hours, err := strconv.Atoi(timeParts[0])
	if err != nil {
        return duration , fmt.Errorf("Invalid hour input: %v", err)
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil {
        return duration , fmt.Errorf("Invalid minute input: %v", err)
	}else if minutes > 60{
        return duration , fmt.Errorf("Invalid minute input: %v", err)
	}

	seconds, err := strconv.Atoi(timeParts[2])
	if err != nil {
		return duration , fmt.Errorf("Invalid second input: %v", err)
	}else if seconds > 60{
		return duration , fmt.Errorf("Invalid second input: %v", err)
	}

	duration = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

    return duration, nil
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

