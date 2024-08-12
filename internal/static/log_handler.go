package static

import (
	"bufio"
	"fmt"
	"log"
	"log_reader/configs"
	"log_reader/internal"
	"log_reader/pkg/logreader"
	"log_reader/pkg/utils"
	static_utils "log_reader/pkg/utils/static"
	"os"
	"strings"
	"sync"
	"time"
)

func ProcessStatic(b *internal.Bot, cfg *configs.Config) {

	duration, err := static_utils.GetTheDuration(cfg.TimeFlag)
	if err != nil {
		log.Fatal("error parsing the time : ", err)
	}

	fmt.Print("phone:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := strings.TrimSpace(scanner.Text())
	phoneNumber:= utils.GetEnteries(line)
	phoneNumber = strings.TrimSpace(phoneNumber)

	if strings.HasPrefix(phoneNumber, "09") {
		phoneNumber = "98" + phoneNumber[1:]
	}
	
	b.InitData(phoneNumber, cfg)

	start := time.Now()

	currentTime := time.Now()
	pastTime := currentTime.Add(-duration)

	clientDirPath := fmt.Sprintf("./logs/%s", cfg.PhoneNumber)
	_, err = os.Stat(clientDirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(clientDirPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	pastTimeStr := pastTime.Format("2006-01-02_15-04-05")
	dirName := fmt.Sprintf("./logs/%s/%s_logs_%s", cfg.PhoneNumber, pastTimeStr, cfg.ClientSystem)
	err = os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	directories := utils.GetLogDirs()
	files := utils.GetLogFiles(directories, cfg)

	fmt.Println("Processing...")
	logreader.ProcessTraces(pastTime, cfg)

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

	static_utils.WriteLogsToFiles(dirName)

	fmt.Printf("Application runtime: %v\n", end)
}
