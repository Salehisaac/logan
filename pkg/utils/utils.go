package utils

import (
	"fmt"
	"log"
	"log_reader/configs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)



func CheckPathValidation(path string)(bool, error){
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

func GetLogDirs() []string {
    directories := []string{"session","status","authsession", "bff", "msg" ,"idgen", "biz", "sync", "media"}

    return directories
}
func GetLogFiles(directories []string, cfg *configs.Config)[]string{
    var logFiles []string

	
	basePath := cfg.LogsPath

	
	for _, dir := range directories {
		accessLogPath := filepath.Join(basePath, dir, "access.log")
		errorLogPath := filepath.Join(basePath, dir, "error.log")
		logFiles = append(logFiles, accessLogPath, errorLogPath)
	}

	return logFiles
}
func CreateDirectories(dirs[]string, base string)error{
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
func ExtractTimeFromLog(line string) (time.Time) {
    var timestamp time.Time
   
    re := regexp.MustCompile(`"@timestamp":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        timestampStr := match[1]
        timestamp, _ = time.Parse(time.RFC3339, timestampStr)
    }

    return timestamp
}
func ExtractContentFromLog(line string) (string) {

    var content string
    re := regexp.MustCompile(`"content":"(.*?)"(?:,|$)`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        content = match[1]
    }

    return content
}
func ExtractTraceFromLog(line string) (string) {

    var trace string
   
    re := regexp.MustCompile(`"trace":"([a-fA-F0-9]{32})"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        trace = match[1]
    }


    return trace
}
func ExtractCallerFromLog(line string) (string) {
   
    var caller string

    re := regexp.MustCompile(`"caller":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        caller = match[1]
    }

    return caller
}
func ExtractLevelFromLog(line string) (string) {
    var  level string

    re := regexp.MustCompile(`"level":"([^"]+)"`)
    match := re.FindStringSubmatch(line)
    if len(match) > 1 {
        level = match[1]
    }
   

    return level
}
func TraceExists(trace string, traces []string) bool {
    for _, t := range traces {
        if t == trace {
            return true
        }
    }
    return false
}


