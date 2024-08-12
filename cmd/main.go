package main

import (
	"flag"
	"log"
	"log_reader/configs"
	bot "log_reader/internal"
	"log_reader/internal/static"
	"log_reader/internal/stream"
	"log_reader/pkg/utils"
)

var (
	streamFlag bool
	timeFlag string
)

func init() {
	flag.BoolVar(&streamFlag, "s", false, "Activate stream mode")
	flag.StringVar(&timeFlag, "time", "", "Time in hr:min:sec format")
}

func main() {

	flag.Parse()

	cfg, err := configs.LoadConfig(streamFlag, timeFlag)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if ok, err := utils.CheckPathValidation(cfg.LogsPath); !ok {
		log.Fatal("Path invalid: ", err)
	}

	b, err := bot.NewBot()
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if cfg.StreamFlag {
		stream.ProcessStream(b, cfg)
	} else {
		static.ProcessStatic(b, cfg)
	}

}
