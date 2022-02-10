package lancache

import (
	"testing"
	"time"
)

func TestLogStatistics_AddEntry(t *testing.T) {
	ls := NewLogCollection()
	e := LogEntry{
		Client:    "steam",
		Src:       "1.2.3.4",
		Timestamp: time.Now(),
		Request:   "GET /something/",
		Size:      123,
		Hit:       false,
		Dest:      "steam.com",
	}
	if ls.alreadyProcessed(e) {
		t.Error("Entry has not already been processed")
	}
	ls.Prepend(&e)
	if len(ls.Logs) != 1 {
		t.Error("Entry was not added properly")
	}
	ls.Prepend(&e)
	if len(ls.Logs) != 1 {
		t.Error("Duplicate entry appears to have been added")
	}
	ee := LogEntry{
		Client:    e.Client,
		Src:       e.Src,
		Timestamp: e.Timestamp,
		Request:   e.Request,
		Size:      e.Size,
		Hit:       e.Hit,
		Dest:      "origin.com",
	}
	ls.Prepend(&ee)
	if len(ls.Logs) != 2 {
		t.Error("Didn't add similar element")
	}
}
