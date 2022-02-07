package lancache

import (
	"log"
	"testing"
)

func TestOpenAccessFileReadOnly(t *testing.T) {
	path := "testdata"
	f, err := openAccessFileReadOnly(path)
	if err != nil {
		t.Error(err)
	}
	if f == nil {
		t.Error("nil file")
	}
}

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
}

func TestParseFileFromScratch(t *testing.T) {
	path := "testdata"
	stats := ParseFileFromScratch(path)
	log.Println(stats.HitEntries)
	log.Println(stats.TotalEntries)
	log.Println(stats.HitBytes)
	log.Println(stats.TotalBytes)
}
