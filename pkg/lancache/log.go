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
	`\"\s[0-9]{3}\s` +
	`(?P<size>[0-9]+)` +
	`\s.+\"(?P<hit>(?:HIT)|(?:MISS)|(?:-))\"\s\"(?P<dest>.+)\"\s\"`)

type LogEntry struct {
	client   string
	source   string
	dateTime time.Time
	request  string
	//code     int64
	byteSize uint64
	hit      bool
	dest     string
}

func (e *LogEntry) GetDateTime() time.Time {
	return e.dateTime
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
	if len(arr) == 0 || len(arr[0]) != 8 {
		log.Println("Unexpected length for parsed regex array; failed regex. Results:")
		log.Println(arr)
		return nil
	}
	t, err := parseDateTime(arr[0][3])
	if err != nil {
		log.Println(err)
	}
	size, err := strconv.ParseUint(arr[0][5], 10, 64)
	if err != nil {
		log.Println(err)
		return nil
	} else if size == 0 {
		log.Println("Not recording entry with bytesize 0")
		return nil
	}
	hit := arr[0][6]
	// healthchecks don't count
	if hit == "-" {
		return nil
	}
	return &LogEntry{
		client:   arr[0][1],
		source:   arr[0][2],
		dateTime: t,
		request:  arr[0][4],
		byteSize: size,
		hit:      hit == "HIT",
		dest:     arr[0][7],
	}
}

// https://go.dev/src/time/format.go have to use these EXACT numbers for the layout
const layout = "02/Jan/2006:15:04:05 -0700"

func parseDateTime(s string) (time.Time, error) {
	return time.Parse(layout, s)
}
