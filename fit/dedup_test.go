package fit

import (
	"testing"
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
)

func newTestRecord(ts time.Time) *mesgdef.Record {
	r := mesgdef.NewRecord(nil)
	r.Timestamp = ts
	return r
}

func TestIsEmptyRecord(t *testing.T) {
	ts := time.Unix(1000, 0)

	empty := newTestRecord(ts)
	if !isEmptyRecord(empty) {
		t.Error("record with no data should be empty")
	}

	withGPS := newTestRecord(ts)
	withGPS.PositionLat = 100
	withGPS.PositionLong = 200
	if isEmptyRecord(withGPS) {
		t.Error("record with GPS data should not be empty")
	}

	withHeartRate := newTestRecord(ts)
	withHeartRate.HeartRate = 150
	if isEmptyRecord(withHeartRate) {
		t.Error("record with heart rate should not be empty")
	}
}

func TestDeduplicate_RemovesEmptyRecords(t *testing.T) {
	d := NewRecordDeduplicator()
	ts := time.Unix(1000, 0)

	result := d.Deduplicate([]*mesgdef.Record{newTestRecord(ts)})
	if len(result) != 0 {
		t.Errorf("expected 0 records, got %d", len(result))
	}
}

func TestDeduplicate_MergesSameTimestamp(t *testing.T) {
	d := NewRecordDeduplicator()
	ts := time.Unix(1000, 0)

	r1 := newTestRecord(ts)
	r1.HeartRate = 150

	r2 := newTestRecord(ts)
	r2.Cadence = 80

	result := d.Deduplicate([]*mesgdef.Record{r1, r2})
	if len(result) != 1 {
		t.Fatalf("expected 1 merged record, got %d", len(result))
	}
	if result[0].HeartRate != 150 {
		t.Errorf("expected merged HeartRate 150, got %d", result[0].HeartRate)
	}
	if result[0].Cadence != 80 {
		t.Errorf("expected merged Cadence 80, got %d", result[0].Cadence)
	}
}

func TestDeduplicate_PreservesDistinctTimestamps(t *testing.T) {
	d := NewRecordDeduplicator()

	r1 := newTestRecord(time.Unix(1000, 0))
	r1.HeartRate = 100
	r2 := newTestRecord(time.Unix(1001, 0))
	r2.HeartRate = 110

	result := d.Deduplicate([]*mesgdef.Record{r1, r2})
	if len(result) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result))
	}
}

func TestIsEmptyValue(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil", nil, true},
		{"invalid uint8", uint8(0xFF), true},
		{"valid uint8", uint8(42), false},
		{"invalid uint16", uint16(0xFFFF), true},
		{"valid uint16", uint16(42), false},
		{"empty string", "", true},
		{"non-empty string", "x", false},
		{"zero time", time.Time{}, true},
		{"non-zero time", time.Unix(1, 0), false},
	}

	for _, c := range cases {
		if got := isEmptyValue(c.value); got != c.want {
			t.Errorf("%s: isEmptyValue(%v) = %v, want %v", c.name, c.value, got, c.want)
		}
	}
}
