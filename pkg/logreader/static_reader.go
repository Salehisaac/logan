package logreader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"log_reader/configs"
	"log_reader/pkg/utils"
	static_utils "log_reader/pkg/utils/static"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const mb = 1024 * 1024
const chunkSize = 1024
var limit int64

func ProcessTraces(pastTime time.Time, cfg *configs.Config) {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	limit_str := os.Getenv("CHUNCK_SIZE_LIMIT")
	limit32 , err := strconv.Atoi(limit_str)
	if err != nil {
		panic(err)
	}
	limitEnv := int64(limit32)
	limit = limitEnv * mb

    dir := filepath.Join(cfg.LogsPath, "session/access.log")
    dir2 := filepath.Join(cfg.LogsPath, "bff/access.log")
    rootDires := []string{dir, dir2}
	var rootWg sync.WaitGroup

    for _, filePath := range rootDires {

	rootWg.Add(1)
	go func(filePath string){
		defer rootWg.Done()
		start_offset, err := findLogOffsetReverse(filePath, pastTime)
		if err != nil {
			panic(err)
		}
        // log.Println("start extracting traces on ", filePath)
        var wg sync.WaitGroup

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

        if start_offset >= fileSize {
            log.Println("Start offset is beyond file size for", filePath)
            return
        }

        
        numChunks := int((fileSize - start_offset) / limit)
        if (fileSize-start_offset)%limit > 0 {
            numChunks++
        }

        for j := 0; j < numChunks; j++ {
            wg.Add(1)
            go func(start int64) {
                defer wg.Done()
                readTraces(start, limit, filePath, pastTime, cfg)
            }(start_offset + int64(j)*limit)
        }

        wg.Wait()
        // log.Println("end extracting traces on ", filePath)
		}(filePath) 
    }
	rootWg.Wait()
}
func readTraces(offset int64, limit int64, fileName string, pastTime time.Time, cfg *configs.Config) {

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

				trace := utils.ExtractTraceFromLog(line)
				content := utils.ExtractContentFromLog(line)

				re := regexp.MustCompile(`perm_auth_key_id:\s*(-?\d+)`)
				match := re.FindStringSubmatch(content)
				if len(match) > 0 {
					if match[1] == cfg.AuthKey {
						mu.Lock()
						if !utils.TraceExists(trace, traces) {
							traces = append(traces, trace)
						}
						mu.Unlock()
					}
				}
			}
		}
	}
}
func ProcessFileLogs(pastTime time.Time, filePath string) {
	// log.Println("start matching logs with traces in ", filePath)
	start_offset, err := findLogOffsetReverse(filePath, pastTime)
		if err != nil {
			// log.Println("didnt find anything in ", filePath)
			return
		}

	var wg sync.WaitGroup

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


	if start_offset >= fileSize {
		log.Println("Start offset is beyond file size for", filePath)
		return
	}

	numChunks := int((fileSize - start_offset) / limit)
        if (fileSize-start_offset)%limit > 0 {
            numChunks++
        }

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(start int64) {
			defer wg.Done()
			static_utils.Read(start, limit, filePath, pastTime, traces)
		}(start_offset + int64(i) * limit)
	}

	wg.Wait()
	// log.Println("end matching logs with traces in ", filePath)
}
func findLogOffsetReverse(filename string, pastTime time.Time) (int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	fileSize := stat.Size()

	var offset int64 = fileSize
	var buffer bytes.Buffer
	var chunk = make([]byte, chunkSize)

	for offset > 0 {
		
		readSize := chunkSize
		if offset < chunkSize {
			readSize = int(offset)
		}

		
		offset -= int64(readSize)

		
		_, err := file.ReadAt(chunk[:readSize], offset)
		if err != nil {
			return 0, err
		}

		
		for i := len(chunk[:readSize]) - 1; i >= 0; i-- {
			buffer.WriteByte(chunk[i])
			if chunk[i] == '\n' || offset == 0 {
				
				line := reverseString(buffer.String())
				buffer.Reset()

				if line != "" {
					timestamp := utils.ExtractTimeFromLog(line)

					if timestamp.Before(pastTime) && !timestamp.IsZero(){
							return offset, nil
						}
				}
			}
		}
	}

	return 0, fmt.Errorf("log not found")
}
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
