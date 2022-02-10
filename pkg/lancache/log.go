package lancache

import (
	"github.com/hpcloud/tail"
	"log"
	"regexp"
	"strconv"
	"sync"
)

var lineRegex = regexp.MustCompile(`^\[(?P<source>.+)\]\s(?P<ip>[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3})\s.+\s[0-9]{3}\s(?P<size>[0-9]+)\s.+\"(?P<hit>(?:HIT)|(?:MISS)|(?:-))\"\s\"(?P<domain>.+)\"\s\"`)

type LogEntry struct {
	source string
	ip     string
	//dateTime string
	//code     int64
	byteSize int64
	hit      bool
	domain   string
}

func ProcessTailAccessFile(tail *tail.Tail, stats *LogStatistics, lock *sync.RWMutex) {
	for line := range tail.Lines {
		if line.Err != nil {
			log.Println(line.Err)
		} else {
			if !stats.AlreadyProcessed(line.Text) {
				entry := ParseLine(line.Text)
				if entry != nil {
					lock.Lock()
					stats.AddEntry(entry, line.Text)
					lock.Unlock()
				}
			}
		}
	}
}

func ParseLine(line string) *LogEntry {
	if line == "" {
		return nil
	}
	arr := lineRegex.FindAllStringSubmatch(line, -1)
	if len(arr) == 0 || len(arr[0]) != 6 {
		log.Println("Unexpected length for parsed regex array; failed regex. Results:")
		log.Println(arr)
		return nil
	}
	size, err := strconv.ParseInt(arr[0][3], 10, 64)
	if err != nil {
		log.Println(err)
		return nil
	}
	// healthchecks don't count
	if arr[0][4] == "-" {
		return nil
	}
	return &LogEntry{
		source:   arr[0][1],
		ip:       arr[0][2],
		byteSize: size,
		hit:      arr[0][4] == "HIT",
		domain:   arr[0][5],
	}
}
