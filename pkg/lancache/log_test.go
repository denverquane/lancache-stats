package lancache

import (
	"github.com/hpcloud/tail"
	"path"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	line := "[steam] 1.2.3.4 / - - - [06/Feb/2022:21:06:20 -0500] \"GET /depot/792101/chunk/12dbb86a0da1552683ed58e3afbdbf0740fb9e24 HTTP/1.1\" 200 1023472 \"-\" \"Valve/Steam HTTP Client 1.0\" \"HIT\" \"edge.steam-dns.top.comcast.net\" \"-\"\n"
	entry := ParseLine(line)
	if entry == nil {
		t.Error("Nil log entry")
	}

	if entry.client != "steam" {
		t.Error("Extracted client incorrectly: " + entry.client)
	}
	if entry.source != "1.2.3.4" {
		t.Error("Extracted IP incorrectly: " + entry.source)
	}
	y, m, d := entry.dateTime.Date()
	if y != 2022 || m != time.February || d != 6 {
		t.Errorf("Incorrect date extracted")
	}
	if entry.dateTime.Hour() != 21 || entry.dateTime.Minute() != 6 || entry.dateTime.Second() != 20 {
		t.Error("Incorrect time extracted")
	}
	if entry.request != "GET /depot/792101/chunk/12dbb86a0da1552683ed58e3afbdbf0740fb9e24 HTTP/1.1" {
		t.Error("Incorrect request extracted")
	}
	if entry.byteSize != 1023472 {
		t.Errorf("Extracted size incorrectly: %d", entry.byteSize)
	}
	if !entry.hit {
		t.Error("Extracted HIT incorrectly (got false)")
	}
	if entry.dest != "edge.steam-dns.top.comcast.net" {
		t.Error("Parsed domain incorrectly: " + entry.dest)
	}
}

func TestParseFileFromScratch(t *testing.T) {
	p := "testdata"
	coll := NewLogCollection()
	tail, err := tail.TailFile(path.Join(p, "access.log"), tail.Config{Follow: false, MustExist: true})
	if err != nil {
		t.Fatal(err)
	}

	ProcessTailAccessFile(tail, &coll)

	if len(coll.data) != 7 {
		t.Error("Expected 7 entries from parsefile")
	}
	summ := coll.Summarize()
	if summ.Hits != 3 {
		t.Error("Expected 3 hit entries from parsefile")
	}
	if summ.HitBytes != 2716352 {
		t.Error("Expected 2716352 hitbytes from parsefile")
	}
	if summ.TotalBytes != 2749082 {
		t.Error("Expected 2749082 total bytes from parsefile")
	}
}

func TestParseDuplicate(t *testing.T) {
	p := "testdata"
	coll := NewLogCollection()
	t1, err := tail.TailFile(path.Join(p, "access.log"), tail.Config{Follow: false, MustExist: true})
	if err != nil {
		t.Fatal(err)
	}
	t2, err := tail.TailFile(path.Join(p, "access.log"), tail.Config{Follow: false, MustExist: true})
	if err != nil {
		t.Fatal(err)
	}

	ProcessTailAccessFile(t1, &coll)
	ProcessTailAccessFile(t2, &coll)
	if len(coll.data) != 7 {
		t.Errorf("Expected 7 entries from parsefile but got %d", len(coll.data))
	}
}
