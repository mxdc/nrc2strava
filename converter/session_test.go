package converter

import (
	"testing"

	"github.com/muktihari/fit/profile/mesgdef"
)

func TestComputeCadenceInSpm(t *testing.T) {
	// 180 steps over 3 minutes = 60 steps/min
	if got := computeCadenceInSpm(180, 3*60*1000); got != 60 {
		t.Errorf("expected 60 spm, got %v", got)
	}
}

func TestComputeCadenceInSpm_ZeroDuration(t *testing.T) {
	if got := computeCadenceInSpm(100, 0); got != 0 {
		t.Errorf("expected 0 for zero duration, got %v", got)
	}
}

func TestComputeMaxSpeed(t *testing.T) {
	r1 := mesgdef.NewRecord(nil)
	r1.SetSpeedScaled(2.0)
	r2 := mesgdef.NewRecord(nil)
	r2.SetSpeedScaled(5.0)
	r3 := mesgdef.NewRecord(nil)
	r3.SetSpeedScaled(3.0)

	if got := computeMaxSpeed([]*mesgdef.Record{r1, r2, r3}); got != 5.0 {
		t.Errorf("expected max speed 5.0, got %v", got)
	}
}

func TestComputeMaxSpeed_EmptyRecords(t *testing.T) {
	if got := computeMaxSpeed([]*mesgdef.Record{}); got != 0 {
		t.Errorf("expected 0 for no records, got %v", got)
	}
}
