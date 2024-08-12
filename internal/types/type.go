package types

import (
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Caller    string
	Content   string
	Level     string
	Trace     string
	FileName  string
}
