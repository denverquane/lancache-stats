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

// TODO test the complexity here to ensure it's actually necessary. Will a regular equality check be faster?
func (s *LogCollection) alreadyProcessed(entry LogEntry) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	if len(s.Logs) == 0 {
		return false
	}
	// if the incoming entry is newer than the first entry, it's certifiably unique
	if entry.Timestamp.After(s.Logs[0].Timestamp) {
		return false

		// when logs come it at identical times
	} else if entry.Timestamp.Equal(s.Logs[0].Timestamp) {
		for _, v := range s.Logs {
			if entry.Timestamp.Equal(v.Timestamp) {
				// an identical element exists already
				if entry == *v {
					return true
				}
			} else {
				// we iterated and found an element where the times are unequal; short-circuit here
				return false
			}
		}
		return false
	} else {
		// incoming entry is older than first entry; by definition we've processed it before
		return true
	}
}

func (s *LogCollection) Prepend(entry *LogEntry) {
	if !s.alreadyProcessed(*entry) {
		s.l.Lock()
		s.Logs = append([]*LogEntry{entry}, s.Logs...)
		s.l.Unlock()
	}
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
