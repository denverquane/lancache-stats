package lancache

import (
	"testing"
	"time"
)

func TestLogStatistics_AddEntry(t *testing.T) {
	ls := NewLogCollection()
	e := LogEntry{
		client:   "steam",
		source:   "1.2.3.4",
		dateTime: time.Now(),
		request:  "GET /something/",
		byteSize: 123,
		hit:      false,
		dest:     "steam.com",
	}
	if ls.alreadyProcessed(e) {
		t.Error("Entry has not already been processed")
	}
	ls.Prepend(&e)
	if len(ls.data) != 1 {
		t.Error("Entry was not added properly")
	}
	ls.Prepend(&e)
	if len(ls.data) != 1 {
		t.Error("Duplicate entry appears to have been added")
	}
	ee := LogEntry{
		client:   e.client,
		source:   e.source,
		dateTime: e.dateTime,
		request:  e.request,
		byteSize: e.byteSize,
		hit:      e.hit,
		dest:     "origin.com",
	}
	ls.Prepend(&ee)
	if len(ls.data) != 2 {
		t.Error("Didn't add similar element")
	}
}
