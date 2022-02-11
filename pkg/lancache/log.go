package lancache

import (
	"github.com/hpcloud/tail"
	"log"
	"regexp"
	"strconv"
	"time"
)

var lineRegex = regexp.MustCompile(`^\[` +
	`(?P<client>.+)` +
	`\]\s` +
	`(?P<src>[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3})` +
	`\s.+\s\[` +
	`(?P<datetime>[0-9]{2}/[a-zA-Z]{3}/[0-9]{4}:.+)` +
	`\]\s\"` +
	`(?P<req>[A-Z]{3,7}\s.+)` +
	`\"\s` +
	`(?P<code>[0-9]{3})` +
	`\s` +
	`(?P<size>[0-9]+)` +
	`\s.+\"(?P<hit>(?:HIT)|(?:MISS)|(?:-))\"\s\"(?P<dest>.+)\"\s\"`)

type LogEntry struct {
	Client    string    `json:"client"`
	Src       string    `json:"src"`
	Timestamp time.Time `json:"timestamp"`
	// can be broken up by HTTP type (GET), URL, and HTTP version
	Request string `json:"request"`
	Code    int64  `json:"code"`
	Size    uint64 `json:"size"`
	Hit     bool   `json:"hit"`
	Dest    string `json:"dest"`
}

func (e1 *LogEntry) Equals(e2 *LogEntry) bool {
	return e1.Client == e2.Client &&
		e1.Src == e2.Src &&
		e1.Timestamp.Equal(e2.Timestamp) &&
		e1.Request == e2.Request &&
		e1.Code == e2.Code &&
		e1.Size == e2.Size &&
		e1.Hit == e2.Hit &&
		e1.Dest == e2.Dest
}

func ProcessTailAccessFile(tail *tail.Tail, col *LogCollection) {
	for line := range tail.Lines {
		if line.Err != nil {
			log.Println(line.Err)
		} else {
			entry := ParseLine(line.Text)
			if entry != nil {
				col.Prepend(entry)
			}
		}
	}
}

func ParseLine(line string) *LogEntry {
	if line == "" {
		return nil
	}
	arr := lineRegex.FindAllStringSubmatch(line, -1)
	if len(arr) == 0 || len(arr[0]) != 9 {
		log.Println("Unexpected length for parsed regex array; failed regex. Results:")
		log.Println(arr)
		return nil
	}
	t, err := parseDateTime(arr[0][3])
	if err != nil {
		log.Println(err)
	}
	code, err := strconv.ParseInt(arr[0][5], 10, 64)
	if err != nil {
		log.Println(err)
		return nil
	}
	size, err := strconv.ParseUint(arr[0][6], 10, 64)
	if err != nil {
		log.Println(err)
		return nil
	} else if size == 0 {
		log.Println("Not recording entry with bytesize 0")
		return nil
	}
	hit := arr[0][7]
	// healthchecks don't count
	if hit == "-" {
		return nil
	}
	return &LogEntry{
		Client:    arr[0][1],
		Src:       arr[0][2],
		Timestamp: t,
		Request:   arr[0][4],
		Code:      code,
		Size:      size,
		Hit:       hit == "HIT",
		Dest:      arr[0][8],
	}
}

// https://go.dev/src/time/format.go have to use these EXACT numbers for the layout
const layout = "02/Jan/2006:15:04:05 -0700"

func parseDateTime(s string) (time.Time, error) {
	return time.Parse(layout, s)
}
