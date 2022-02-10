package ttl

import (
	"testing"
	"time"
)

type dateTimeStruct struct {
	dateTime time.Time
}

func (d *dateTimeStruct) GetDateTime() time.Time {
	return d.dateTime
}

func TestTTLSlice_Prepend(t *testing.T) {
	s := NewTTLSlice(time.Second, time.Hour, make([]HasDateTime, 0))
	e := dateTimeStruct{
		dateTime: time.Now(),
	}
	err := s.Add(&e)
	if err != nil {
		t.Error(err)
	}
	newEntry := dateTimeStruct{
		dateTime: e.dateTime.Add(time.Second),
	}
	err = s.Add(&newEntry)
	if err != nil {
		t.Error(err)
	}

	oldEntry := dateTimeStruct{
		dateTime: newEntry.dateTime.Add(-time.Second),
	}
	err = s.Add(&oldEntry)
	if err == nil {
		t.Error("Expected error when adding old entry to TTLSlice, but no error received")
	}
}

func TestTTLSlice_Expiry(t *testing.T) {
	s := NewTTLSlice(time.Millisecond, time.Millisecond, make([]HasDateTime, 0))
	e := dateTimeStruct{
		dateTime: time.Now(),
	}
	err := s.Add(&e)
	if err != nil {
		t.Error(err)
	}
	// give the worker time to remove it
	time.Sleep(time.Millisecond * 2)

	elems := s.GetAll()
	if len(elems) != 0 {
		t.Errorf("Expected all data to have expired, but an element is still present")
	}
	e = dateTimeStruct{
		dateTime: time.Now().Add(time.Minute), // just make sure the data can't be expired in any situation
	}
	err = s.Add(&e)
	if err != nil {
		t.Error(err)
	}
	// give the worker time to NOT remove it
	time.Sleep(time.Millisecond * 2)

	elems = s.GetAll()
	if len(elems) != 1 {
		t.Errorf("Expected no elements to have expired")
	}
}
