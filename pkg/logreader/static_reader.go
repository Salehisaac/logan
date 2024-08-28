package logreader

import (
	"bufio"
	"io"
	"log_reader/configs"
	"log_reader/pkg/utils"
	static_utils "log_reader/pkg/utils/static"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const mb = 1024 * 1024

func ProcessTraces(pastTime time.Time, cfg *configs.Config) {

	dir := filepath.Join(cfg.LogsPath, "session/access.log")
	dir2 := filepath.Join(cfg.LogsPath, "bff/access.log")
	rootDires := []string{dir, dir2}

	for _, filePath := range rootDires {
		var wg sync.WaitGroup

		var limit int64 = 100 * mb

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
				readTraces(start, limit, filePath, pastTime, cfg)
			}(int64(i) * limit)
		}

		wg.Wait()
	}
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
	var wg sync.WaitGroup

	var limit int64 = 100 * mb

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
			static_utils.Read(start, limit, filePath, pastTime, traces)
		}(int64(i) * limit)
	}

	wg.Wait()
}
