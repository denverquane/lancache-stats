package lancache

import (
	"bufio"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
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

func openAccessFileReadOnly(logpath string) (*os.File, error) {
	return os.OpenFile(path.Join(logpath, "access.log"), os.O_RDONLY, 0666)
}

func ParseFileFromOffset(logpath string, stats *LogStatistics, offset int64) int64 {
	f, err := openAccessFileReadOnly(logpath)
	if err != nil {
		log.Println(err)
		return -1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	pos := offset
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		pos += int64(advance)
		return
	}
	scanner.Split(scanLines)
	for scanner.Scan() {
		entry := ParseLine(scanner.Text())
		if entry != nil {
			stats.AddEntry(entry)
		}
	}
	return pos
}

func ParseLine(line string) *LogEntry {
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

func LogFileSize(logpath string) (int64, error) {
	f, err := openAccessFileReadOnly(logpath)
	if err != nil {
		return -1, err
	}
	fi, err := f.Stat()
	if err != nil {
		return -1, err
	}
	return fi.Size(), nil
}
