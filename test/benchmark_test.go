package logreader

import (
	"fmt"
	"log_reader/configs"
	"log_reader/pkg/utils"
	"log_reader/pkg/logreader"
	"log_reader/pkg/utils/static"
	"os"
	"sync"
	"testing"
	"time"
)

func BenchmarkStaticMode(b *testing.B) {
	cfg := configs.Config{
		LogsPath:    "/opt/test/teamgram-server/logs",
		ClientSystem: "80L0_Windows 10_172.20.0.1",
		AuthKey:      "5145167012953315267",
		StreamFlag:   false,
		PhoneNumber:  "989172568979 ",
		TimeFlag:     "1:00:00",
	}
	duration, err := static.GetTheDuration(cfg.TimeFlag)
	if err != nil {
		b.Fatal("error parsing the time: ", err)
	}

	start := time.Now()

	currentTime := time.Now()
	pastTime := currentTime.Add(-duration)

	clientDirPath := fmt.Sprintf("./logs/%s", cfg.PhoneNumber)
	_, err = os.Stat(clientDirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(clientDirPath, os.ModePerm)
		if err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}
	}

	pastTimeStr := pastTime.Format("2006-01-02_15-04-05")
	dirName := fmt.Sprintf("./logs/%s/%s_logs_%s", cfg.PhoneNumber, pastTimeStr, cfg.ClientSystem)
	err = os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		b.Fatalf("Failed to create directory: %v", err)
	}

	directories := utils.GetLogDirs()
	files := utils.GetLogFiles(directories, &cfg)

	b.Log("Processing...")
	logreader.ProcessTraces(pastTime, &cfg)

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			logreader.ProcessFileLogs(pastTime, file)
		}(file)
	}
	wg.Wait()
	end := time.Since(start)

	static.WriteLogsToFiles(dirName)

	b.Logf("Application runtime: %v\n", end)
}

//go test -bench=. -benchmem -count=10 |tee static.txt

