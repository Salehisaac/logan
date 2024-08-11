package stream

import(
    "fmt"
	"bufio"
	"strings"
	"os"
	"log_reader/configs"
	"log_reader/internal"
	stream_utils "log_reader/pkg/utils/stream"
    "log_reader/pkg/logreader"	
)


func ProcessStream(b *internal.Bot , cfg *configs.Config ){
	fmt.Print("phone: ")
        scanner := bufio.NewScanner(os.Stdin)
        scanner.Scan()
        line := strings.TrimSpace(scanner.Text())
        phoneNumber := stream_utils.GetEnteries(line)
        phoneNumber = strings.TrimSpace(phoneNumber)
        if strings.HasPrefix(phoneNumber, "09") {
            phoneNumber = "98" + phoneNumber[1:]
        }
        b.InitData(phoneNumber, cfg)
        logreader.StartStream(cfg)
}



