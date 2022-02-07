package lancache

import (
	"bufio"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
)

var lineRegex = regexp.MustCompile(`^\[(?P<source>.+)\]\s(?P<ip>[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3})\s.+\s[0-9]{3}\s(?P<size>[0-9]+)\s.+\"(?P<hit>(?:HIT)|(?:MISS))\"`)

type LogStatistics struct {
	HitBytes     uint64
	TotalBytes   uint64
	HitEntries   uint64
	TotalEntries uint64
}

type LogEntry struct {
	source string
	ip     string
	//dateTime string
	//code     int64
	byteSize int64
	hit      bool
}

func openAccessFileReadOnly(logpath string) (*os.File, error) {
	return os.OpenFile(path.Join(logpath, "access.log"), os.O_RDONLY, 0666)
}

func ParseFileFromScratch(logpath string) LogStatistics {
	f, err := openAccessFileReadOnly(logpath)
	if err != nil {
		log.Println(err)
		return LogStatistics{}
	}
	defer f.Close()

	var stats LogStatistics
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		entry := ParseLine(scanner.Text())
		if entry != nil {
			stats.TotalBytes += uint64(entry.byteSize)
			if entry.hit {
				stats.HitEntries++
				stats.HitBytes += uint64(entry.byteSize)
			}
		}
	}
	return stats
}

func ParseLine(line string) *LogEntry {
	arr := lineRegex.FindAllStringSubmatch(line, -1)
	if len(arr) == 0 || len(arr[0]) != 5 {
		log.Println("Unexpected length for parsed regex array; failed regex. Results:")
		log.Println(arr)
		return nil
	}
	size, err := strconv.ParseInt(arr[0][3], 10, 64)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &LogEntry{
		source:   arr[0][1],
		ip:       arr[0][2],
		byteSize: size,
		hit:      arr[0][4] == "HIT",
	}
}
