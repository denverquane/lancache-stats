package ttl

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type HasDateTime interface {
	GetDateTime() time.Time
}

type TTLSlice struct {
	data []HasDateTime
	l    sync.Mutex
}

func NewTTLSlice(tick time.Duration, ttl time.Duration, data []HasDateTime) *TTLSlice {
	s := TTLSlice{
		data: data,
		l:    sync.Mutex{},
	}
	go func() {
		for now := range time.Tick(tick) {
			s.l.Lock()

			for i, entry := range s.data {
				if now.Add(-ttl).After(entry.GetDateTime()) {
					// cut off all remaining entries; if this one is expired, so are all the following ones
					s.data = s.data[:i]
					break
				}
			}

			s.l.Unlock()
		}
	}()
	return &s
}

func (s *TTLSlice) Add(entry HasDateTime) error {
	s.l.Lock()
	defer s.l.Unlock()

	if len(s.data) == 0 {
		s.data = append([]HasDateTime{entry}, s.data...)
		return nil
	}

	if entry.GetDateTime().Before(s.data[0].GetDateTime()) {
		return errors.New(fmt.Sprintf("cant prepend entry w/ datetime %s to slice w/ first datetime %s", entry.GetDateTime().String(), s.data[0].GetDateTime()))
	}
	s.data = append([]HasDateTime{entry}, s.data...)
	return nil
}

func (s *TTLSlice) GetAll() []HasDateTime {
	s.l.Lock()
	defer s.l.Unlock()

	return s.data
}
