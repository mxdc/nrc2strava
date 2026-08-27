package converter

import (
	"math"
	"testing"
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/mxdc/nrc2strava/types"
)

func newRecordAt(ts time.Time) *mesgdef.Record {
	r := mesgdef.NewRecord(nil)
	r.SetTimestamp(ts)
	return r
}

func TestInterpolateValue_BeforeFirstInterval(t *testing.T) {
	metric := types.Metric{
		Values: []types.MetricValue{
			{StartEpochMs: 10000, EndEpochMs: 20000, Value: 5.0},
		},
	}
	if got := interpolateValue(5, metric); got != 5.0 {
		t.Errorf("expected 5.0, got %v", got)
	}
}

func TestInterpolateValue_WithinInterval(t *testing.T) {
	metric := types.Metric{
		Values: []types.MetricValue{
			{StartEpochMs: 0, Value: 0.0},
			{StartEpochMs: 10000, Value: 10.0},
		},
	}
	// Halfway between 0s and 10s should interpolate to 5.0
	if got := interpolateValue(5, metric); got != 5.0 {
		t.Errorf("expected 5.0, got %v", got)
	}
}

func TestInterpolateValue_AfterLastInterval(t *testing.T) {
	metric := types.Metric{
		Values: []types.MetricValue{
			{StartEpochMs: 0, Value: 1.0},
			{StartEpochMs: 10000, Value: 2.0},
		},
	}
	if got := interpolateValue(20, metric); got != 2.0 {
		t.Errorf("expected 2.0, got %v", got)
	}
}

func TestConvertStepsToCadence(t *testing.T) {
	// 60 steps over 30 seconds = 120 steps/min = 60 rpm (2 steps/rev)
	stepsMetric := types.Metric{
		Values: []types.MetricValue{
			{StartEpochMs: 0, EndEpochMs: 30000, Value: 60},
		},
	}

	cadence := convertStepsToCadence(stepsMetric)
	if len(cadence.Values) != 1 {
		t.Fatalf("expected 1 cadence value, got %d", len(cadence.Values))
	}
	if got := cadence.Values[0].Value; got != 60 {
		t.Errorf("expected 60 rpm, got %v", got)
	}
}

func TestConvertStepsToCadence_ClampsAbove180Rpm(t *testing.T) {
	// An implausibly high step rate should be clamped to 0, not reported
	stepsMetric := types.Metric{
		Values: []types.MetricValue{
			{StartEpochMs: 0, EndEpochMs: 1000, Value: 1000},
		},
	}

	cadence := convertStepsToCadence(stepsMetric)
	if got := cadence.Values[0].Value; got != 0 {
		t.Errorf("expected clamped rpm to be 0, got %v", got)
	}
}

func TestFillElevation_InterpolatesBetweenPoints(t *testing.T) {
	base := time.Unix(1000, 0)
	records := []*mesgdef.Record{
		newRecordAt(base),
		newRecordAt(base.Add(5 * time.Second)),
		newRecordAt(base.Add(10 * time.Second)),
	}

	elevationMetric := types.Metric{
		Type: "elevation",
		Values: []types.MetricValue{
			{StartEpochMs: base.Unix() * 1000, Value: 100},
			{StartEpochMs: (base.Unix() + 10) * 1000, Value: 200},
		},
	}

	fillElevation(records, elevationMetric)

	if got := records[0].AltitudeScaled(); got != 100 {
		t.Errorf("expected 100 at start, got %v", got)
	}
	if got := records[1].AltitudeScaled(); got != 150 {
		t.Errorf("expected 150 at midpoint, got %v", got)
	}
	if got := records[2].AltitudeScaled(); got != 200 {
		t.Errorf("expected 200 at end, got %v", got)
	}
}

func TestFillElevation_NoOpWhenWrongMetricType(t *testing.T) {
	records := []*mesgdef.Record{newRecordAt(time.Unix(1000, 0))}

	fillElevation(records, types.Metric{Type: "distance", Values: []types.MetricValue{{Value: 42}}})

	// An unset altitude decodes as NaN; fillElevation should leave it untouched
	// rather than setting it from a metric of the wrong type.
	if got := records[0].AltitudeScaled(); !math.IsNaN(got) {
		t.Errorf("expected altitude to remain unset (NaN) for wrong metric type, got %v", got)
	}
}
