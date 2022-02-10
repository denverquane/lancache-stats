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

	if entry.Client != "steam" {
		t.Error("Extracted client incorrectly: " + entry.Client)
	}
	if entry.Src != "1.2.3.4" {
		t.Error("Extracted IP incorrectly: " + entry.Src)
	}
	y, m, d := entry.Timestamp.Date()
	if y != 2022 || m != time.February || d != 6 {
		t.Errorf("Incorrect date extracted")
	}
	if entry.Timestamp.Hour() != 21 || entry.Timestamp.Minute() != 6 || entry.Timestamp.Second() != 20 {
		t.Error("Incorrect time extracted")
	}
	if entry.Request != "GET /depot/792101/chunk/12dbb86a0da1552683ed58e3afbdbf0740fb9e24 HTTP/1.1" {
		t.Error("Incorrect request extracted")
	}
	if entry.Size != 1023472 {
		t.Errorf("Extracted size incorrectly: %d", entry.Size)
	}
	if !entry.Hit {
		t.Error("Extracted HIT incorrectly (got false)")
	}
	if entry.Dest != "edge.steam-dns.top.comcast.net" {
		t.Error("Parsed domain incorrectly: " + entry.Dest)
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

	if len(coll.Logs) != 7 {
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
	if len(coll.Logs) != 7 {
		t.Errorf("Expected 7 entries from parsefile but got %d", len(coll.Logs))
	}
}
