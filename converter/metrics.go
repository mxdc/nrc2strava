package converter

import (
	"strings"
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/mxdc/nrc2strava/types"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// MetricsConverter converts the NIke Activities Metrics into FIT Activity Records
type MetricsConverter struct {
	StartEpochMs     int64
	EndEpochMs       int64
	ActiveDurationMs int64

	// Outdoor or Treadmill
	Indoor bool

	// Raw data
	Moments   []types.Moment
	Metrics   []types.Metric
	Summaries []types.Summary

	// Raw Metrics
	DistanceMetric  types.Metric
	LongitudeMetric types.Metric
	LatitudeMetric  types.Metric
	AscentMetric    types.Metric
	DescentMetric   types.Metric
	ElevationMetric types.Metric
	StepsMetric     types.Metric
	SpeedMetric     types.Metric
	PaceMetric      types.Metric

	// Summaries
	SpeedSummary    types.Summary
	StepsSummary    types.Summary
	AscentSummary   types.Summary
	CaloriesSummary types.Summary
	PaceSummary     types.Summary
	DistanceSummary types.Summary
	DescentSummary  types.Summary
	NikefuelSummary types.Summary

	// logger
	logger *logrus.Logger
}

func isIndoor(tags map[string]string) bool {
	if location, ok := tags["location"]; ok {
		return strings.HasPrefix(strings.ToLower(location), "indoor")
	}

	return false
}

// NewMetricsConverter returns an initialized MetricsConverter
func NewMetricsConverter(
	StartEpochMs int64,
	EndEpochMs int64,
	ActiveDurationMs int64,
	Metrics []types.Metric,
	Summaries []types.Summary,
	Moments []types.Moment,
	Tags map[string]string,
) *MetricsConverter {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	parser := &MetricsConverter{
		StartEpochMs:     StartEpochMs,
		EndEpochMs:       EndEpochMs,
		ActiveDurationMs: ActiveDurationMs,
		Indoor:           isIndoor(Tags),
		Moments:          Moments,
		Metrics:          Metrics,
		Summaries:        Summaries,
		logger:           logger,
	}

	for _, summary := range Summaries {
		if summary.Metric == "steps" {
			parser.StepsSummary = summary
		}
		if summary.Metric == "speed" {
			parser.SpeedSummary = summary
		}
		if summary.Metric == "nikefuel" {
			parser.NikefuelSummary = summary
		}
		if summary.Metric == "ascent" {
			parser.AscentSummary = summary
		}
		if summary.Metric == "calories" {
			parser.CaloriesSummary = summary
		}
		if summary.Metric == "pace" {
			parser.PaceSummary = summary
		}
		if summary.Metric == "distance" {
			parser.DistanceSummary = summary
		}
		if summary.Metric == "descent" {
			parser.DescentSummary = summary
		}
	}

	for _, metric := range Metrics {
		if metric.Type == "distance" {
			parser.DistanceMetric = metric
		}
		if metric.Type == "latitude" {
			parser.LatitudeMetric = metric
		}
		if metric.Type == "longitude" {
			parser.LongitudeMetric = metric
		}
		if metric.Type == "ascent" {
			parser.AscentMetric = metric
		}
		if metric.Type == "descent" {
			parser.DescentMetric = metric
		}
		if metric.Type == "elevation" {
			parser.ElevationMetric = metric
		}
		if metric.Type == "steps" {
			parser.StepsMetric = metric
		}
		if metric.Type == "speed" {
			parser.SpeedMetric = metric
		}
		if metric.Type == "pace" {
			parser.PaceMetric = metric
		}
	}
	return parser
}

func (m *MetricsConverter) ParseRecords() []*mesgdef.Record {
	// Convert milliseconds to seconds
	StartEpochSeconds := m.StartEpochMs / 1000
	EndEpochSeconds := m.EndEpochMs / 1000
	totalRecords := EndEpochSeconds - StartEpochSeconds + 1
	m.logger.Debugf("Number of records: %d\n", totalRecords)

	records := make([]*mesgdef.Record, totalRecords)
	for i := range totalRecords {
		timestamp := time.Unix(StartEpochSeconds+i, 0).UTC()
		records[i] = mesgdef.NewRecord(nil)
		records[i].SetTimestamp(timestamp)
	}

	fillCadenceFromSteps(records, m.StepsMetric)
	fillDistance(records, m.DistanceMetric)
	fillPositionFromGPS(records, m.LatitudeMetric, m.LongitudeMetric)
	fillElevation(records, m.ElevationMetric)
	fillSpeedFromDistance(records)

	return records
}
