package stream

import (
	"bufio"
	"fmt"
	"log_reader/configs"
	"log_reader/internal"
	"log_reader/pkg/logreader"
	"log_reader/pkg/utils"
	"os"
	"strings"
)

func ProcessStream(b *internal.Bot, cfg *configs.Config) {
	fmt.Print("phone: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := strings.TrimSpace(scanner.Text())
	phoneNumber := utils.GetEnteries(line)
	phoneNumber = strings.TrimSpace(phoneNumber)
	if strings.HasPrefix(phoneNumber, "09") {
		phoneNumber = "98" + phoneNumber[1:]
	}
	b.InitData(phoneNumber, cfg)
	logreader.StartStream(cfg)
}
