package lancache

import "sync"

type LogCollection struct {
	data []*LogEntry
	l    sync.RWMutex
}

func NewLogCollection() LogCollection {
	return LogCollection{data: make([]*LogEntry, 0)}
}

func (s *LogCollection) alreadyProcessed(entry LogEntry) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	if len(s.data) == 0 {
		return false
	}
	// if the incoming entry is newer than the first entry, it's certifiably unique
	if entry.dateTime.After(s.data[0].dateTime) {
		return false
	} else if entry.dateTime.Equal(s.data[0].dateTime) {
		for _, v := range s.data {
			if entry.dateTime.Equal(v.dateTime) {
				if entry == *v {
					return true
				}
				// the else case is one where the times match, but the lines are different
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
		s.data = append([]*LogEntry{entry}, s.data...)
		s.l.Unlock()
	}
}

func (s *LogCollection) Summarize() (c CacheRecord) {
	s.l.RLock()
	for _, v := range s.data {
		c.Total++
		c.TotalBytes += v.byteSize
		if v.hit {
			c.Hits++
			c.HitBytes += v.byteSize
		}
	}
	s.l.RUnlock()
	return c
}

type CacheRecord struct {
	Hits       uint64 `json:"hit"`
	Total      uint64 `json:"total"`
	HitBytes   uint64 `json:"hit_bytes"`
	TotalBytes uint64 `json:"total_bytes"`
}
