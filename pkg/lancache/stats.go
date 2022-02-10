package lancache

type EntryAddable interface {
	AddEntry(entry *LogEntry)
}

type Domains map[string]CacheRecord

type LogStatistics struct {
	Summary      CacheRecord               `json:"summary"`
	Domains      Domains                   `json:"domains"`
	Requests     map[string]RequesterStats `json:"requests"`
	rawStringMap map[string]struct{}
}

func NewLogStatistics() LogStatistics {
	return LogStatistics{
		Summary:      EmptyCacheRecord(),
		Domains:      make(Domains),
		Requests:     make(map[string]RequesterStats),
		rawStringMap: map[string]struct{}{},
	}
}

func (s *LogStatistics) AlreadyProcessed(rawText string) bool {
	_, ok := s.rawStringMap[rawText]
	return ok
}

func (s *LogStatistics) AddEntry(entry *LogEntry, rawText string) {
	s.Summary.AddEntry(entry)

	if d, ok := s.Domains[entry.domain]; !ok {
		s.Domains[entry.domain] = NewCacheRecord(entry)
	} else {
		d.AddEntry(entry)
		s.Domains[entry.domain] = d
	}

	if i, ok := s.Requests[entry.ip]; !ok {
		s.Requests[entry.ip] = NewIpStat(entry)
	} else {
		i.AddEntry(entry)
		s.Requests[entry.ip] = i
	}
	s.rawStringMap[rawText] = struct{}{}
}

type CacheRecord struct {
	Hits       uint64 `json:"hit"`
	Total      uint64 `json:"total"`
	HitBytes   uint64 `json:"hit_bytes"`
	TotalBytes uint64 `json:"total_bytes"`
}

func (c *CacheRecord) AddEntry(entry *LogEntry) {
	c.TotalBytes += uint64(entry.byteSize)
	c.Total++
	if entry.hit {
		c.HitBytes += uint64(entry.byteSize)
		c.Hits++
	}
}

func EmptyCacheRecord() CacheRecord {
	return CacheRecord{}
}

func NewCacheRecord(entry *LogEntry) CacheRecord {
	c := CacheRecord{
		HitBytes:   0,
		TotalBytes: uint64(entry.byteSize),
		Hits:       0,
		Total:      1,
	}
	if entry.hit {
		c.HitBytes = uint64(entry.byteSize)
		c.Hits = 1
	}
	return c
}

type RequesterStats struct {
	CanonicalName string      `json:"canonical_name"`
	Summary       CacheRecord `json:"summary"`
	Domains       Domains     `json:"domains"`
}

func NewIpStat(entry *LogEntry) RequesterStats {
	domains := make(Domains)
	domains[entry.domain] = NewCacheRecord(entry)
	return RequesterStats{
		CanonicalName: "",
		Summary:       NewCacheRecord(entry),
		Domains:       domains,
	}
}

func (i *RequesterStats) AddEntry(entry *LogEntry) {
	i.Summary.AddEntry(entry)
	if d, dok := i.Domains[entry.domain]; !dok {
		i.Domains[entry.domain] = NewCacheRecord(entry)
	} else {
		d.AddEntry(entry)
		i.Domains[entry.domain] = d
	}
}
