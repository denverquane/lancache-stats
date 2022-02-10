package lancache

import (
	"github.com/hpcloud/tail"
	"path"
	"sync"
	"testing"
)

func TestParseLine(t *testing.T) {
	line := "[steam] 1.2.3.4 / - - - [06/Feb/2022:21:06:20 -0500] \"GET /depot/792101/chunk/12dbb86a0da1552683ed58e3afbdbf0740fb9e24 HTTP/1.1\" 200 1023472 \"-\" \"Valve/Steam HTTP Client 1.0\" \"HIT\" \"edge.steam-dns.top.comcast.net\" \"-\"\n"
	entry := ParseLine(line)
	if entry == nil {
		t.Error("Nil log entry")
	}

	if entry.source != "steam" {
		t.Error("Extracted source incorrectly: " + entry.source)
	}
	if entry.ip != "1.2.3.4" {
		t.Error("Extracted IP incorrectly: " + entry.ip)
	}
	if entry.byteSize != 1023472 {
		t.Errorf("Extracted size incorrectly: %d", entry.byteSize)
	}
	if !entry.hit {
		t.Error("Extracted HIT incorrectly (got false)")
	}
	if entry.domain != "edge.steam-dns.top.comcast.net" {
		t.Error("Parsed domain incorrectly: " + entry.domain)
	}
}

func TestParseFileFromScratch(t *testing.T) {
	p := "testdata"
	stats := NewLogStatistics()
	tail, err := tail.TailFile(path.Join(p, "access.log"), tail.Config{Follow: false, MustExist: true})
	if err != nil {
		t.Fatal(err)
	}
	lock := sync.RWMutex{}

	ProcessTailAccessFile(tail, &stats, &lock)

	if stats.Summary.Total != 7 {
		t.Error("Expected 7 entries from parsefile")
	}
	if stats.Summary.Hits != 3 {
		t.Error("Expected 3 hit entries from parsefile")
	}
	if stats.Summary.HitBytes != 2716352 {
		t.Error("Expected 2716352 hitbytes from parsefile")
	}
	if stats.Summary.TotalBytes != 2749082 {
		t.Error("Expected 2749082 total bytes from parsefile")
	}

	if len(stats.Requests) != 2 {
		t.Error("Expected 2 ips in stats")
	}
	good := stats.Requests["1.2.3.4"].Summary
	if good.TotalBytes != good.HitBytes {
		t.Error("Hit and total bytes should be equal for 1.2.3.4")
	}
	if good.Hits != good.Total {
		t.Error("Hit and total entries should be equal for 1.2.3.4")
	}

	bad := stats.Requests["4.3.2.1"].Summary
	if bad.Hits != 0 {
		t.Errorf("4.3.2.1 should have 0 hits but has %d", bad.Hits)
	}
	if bad.HitBytes != 0 {
		t.Errorf("4.3.2.1 should have 0 hit bytes but has %d", bad.HitBytes)
	}
	if bad.TotalBytes != 32730 {
		t.Errorf("4.3.2.1 should have 32730 total bytes but has %d", bad.TotalBytes)
	}
	if bad.Total != 4 {
		t.Errorf("4.3.2.1 should have 4 total entries but has %d", bad.Total)
	}
}
