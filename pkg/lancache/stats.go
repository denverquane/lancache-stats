package lancache

import (
	"sync"
	"time"
)

type LogCollection struct {
	Logs []*LogEntry `json:"logs"`
	l    sync.RWMutex
}

func NewLogCollection() LogCollection {
	return LogCollection{Logs: make([]*LogEntry, 0)}
}

func (s *LogCollection) alreadyProcessed(entry LogEntry) bool {
	for _, v := range s.Logs {
		if entry.Equals(v) {
			return true
		}
	}
	return false
}

func (s *LogCollection) Prepend(entry *LogEntry) {
	s.l.Lock()
	if !s.alreadyProcessed(*entry) {
		s.Logs = append([]*LogEntry{entry}, s.Logs...)
	}
	s.l.Unlock()
}

type LogSummary struct {
	Hits       uint64 `json:"hit"`
	Total      uint64 `json:"total"`
	HitBytes   uint64 `json:"hit_bytes"`
	TotalBytes uint64 `json:"total_bytes"`
}

func (s *LogCollection) Summarize() (c LogSummary) {
	s.l.RLock()
	for _, v := range s.Logs {
		c.Total++
		c.TotalBytes += v.Size
		if v.Hit {
			c.Hits++
			c.HitBytes += v.Size
		}
	}
	s.l.RUnlock()
	return c
}

// LogPredicateFn is a function that tests a LogEntry, and returns t/f if it satisfies the predicate
type LogPredicateFn func(LogEntry) bool

// Filter tests all elements of the LogCollection against a variadic number of filter predicates.
// It returns the subset of elements that fulfill all predicates passed to it
func (s *LogCollection) Filter(preds []LogPredicateFn) *LogCollection {
	cc := NewLogCollection()
	s.l.RLock()
	for _, v := range s.Logs {
		filter := false
		for _, pred := range preds {
			if !pred(*v) {
				filter = true
				break
			}
		}
		if !filter {
			cc.Logs = append(cc.Logs, v)
		}
	}
	s.l.RUnlock()
	return &cc
}

// ClientPredFn returns a predicate function that tests for the LogEntry having a matching client
func ClientPredFn(client string) LogPredicateFn {
	return func(entry LogEntry) bool {
		return client == "" || client == entry.Client
	}
}

// SrcPredFn returns a predicate function that tests for the LogEntry having a matching src
func SrcPredFn(src string) LogPredicateFn {
	return func(entry LogEntry) bool {
		return src == "" || src == entry.Src
	}
}

// DestPredFn returns a predicate function that tests for the LogEntry having a matching dest
func DestPredFn(dest string) LogPredicateFn {
	return func(entry LogEntry) bool {
		return dest == "" || dest == entry.Dest
	}
}

func TimeRangePredFn(start, end time.Time) LogPredicateFn {
	return func(entry LogEntry) bool {
		return entry.Timestamp.After(start) && entry.Timestamp.Before(end)
	}
}

func SizeRangePredFn(min, max uint64) LogPredicateFn {
	return func(entry LogEntry) bool {
		return entry.Size > min && entry.Size < max
	}
}
