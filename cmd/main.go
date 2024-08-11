package main

import(
	"flag"
	"log"
	"log_reader/configs"
	"log_reader/pkg/utils"
	bot "log_reader/internal"
    "log_reader/internal/stream"
    "log_reader/internal/static"
)
var(
	streamFlag bool
)

func init() {
    flag.BoolVar(&streamFlag, "s", false, "Activate stream mode")
}

func main(){
	flag.Parse()

	cfg, err := configs.LoadConfig(streamFlag)
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

