package converter

import (
	"strings"
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
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

// InitMetricsConverter returns an initialized MetricsConverter
func InitMetricsConverter(
	StartEpochMs int64,
	EndEpochMs int64,
	ActiveDurationMs int64,
	Metrics []types.Metric,
	Summaries []types.Summary,
	Moments []types.Moment,
	Tags map[string]string,
) *MetricsConverter {
	var parser MetricsConverter

	parser.logger = logrus.New()
	parser.logger.SetFormatter(utils.LogFormat)

	parser.StartEpochMs = StartEpochMs
	parser.EndEpochMs = EndEpochMs
	parser.Indoor = isIndoor(Tags)
	parser.Moments = Moments
	parser.ActiveDurationMs = ActiveDurationMs
	parser.Metrics = Metrics
	parser.Summaries = Summaries

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
	return &parser
}

func interpolateLatitude(timestamp int64, latitudeMetric types.Metric) float64 {
	// Handle timestamps before the first interval
	if len(latitudeMetric.Values) > 0 {
		first := latitudeMetric.Values[0]
		firstStartSeconds := first.StartEpochMs / 1000
		if timestamp < firstStartSeconds {
			return first.Value
		}
	}

	// Iterate through latitude intervals
	for i := 0; i < len(latitudeMetric.Values)-1; i++ {
		current := latitudeMetric.Values[i]
		next := latitudeMetric.Values[i+1]

		// Convert start and end times to seconds
		currentStartSeconds := current.StartEpochMs / 1000
		nextStartSeconds := next.StartEpochMs / 1000

		// Check if the timestamp falls within the current interval
		if timestamp >= currentStartSeconds && timestamp < nextStartSeconds {
			intervalDuration := float64(nextStartSeconds - currentStartSeconds)
			timeElapsed := float64(timestamp - currentStartSeconds)
			latitudeDelta := next.Value - current.Value
			return current.Value + (latitudeDelta * (timeElapsed / intervalDuration))
		}
	}

	// Handle timestamps after the last interval
	last := latitudeMetric.Values[len(latitudeMetric.Values)-1]
	if timestamp >= last.StartEpochMs/1000 {
		return last.Value
	}

	return 0 // Default value if no match is found
}

func interpolateLongitude(timestamp int64, longitudeMetric types.Metric) float64 {
	// Handle timestamps before the first interval
	if len(longitudeMetric.Values) > 0 {
		first := longitudeMetric.Values[0]
		firstStartSeconds := first.StartEpochMs / 1000
		if timestamp < firstStartSeconds {
			return first.Value
		}
	}

	// Iterate through longitude intervals
	for i := 0; i < len(longitudeMetric.Values)-1; i++ {
		current := longitudeMetric.Values[i]
		next := longitudeMetric.Values[i+1]

		// Convert start and end times to seconds
		currentStartSeconds := current.StartEpochMs / 1000
		nextStartSeconds := next.StartEpochMs / 1000

		// Check if the timestamp falls within the current interval
		if timestamp >= currentStartSeconds && timestamp < nextStartSeconds {
			intervalDuration := float64(nextStartSeconds - currentStartSeconds)
			timeElapsed := float64(timestamp - currentStartSeconds)
			longitudeDelta := next.Value - current.Value
			return current.Value + (longitudeDelta * (timeElapsed / intervalDuration))
		}
	}

	// Handle timestamps after the last interval
	last := longitudeMetric.Values[len(longitudeMetric.Values)-1]
	if timestamp >= last.StartEpochMs/1000 {
		return last.Value
	}

	return 0 // Default value if no match is found
}

func fillPositionFromGPS(records []*mesgdef.Record, latitudeMetric, longitudeMetric types.Metric) {
	if latitudeMetric.Type != "latitude" || longitudeMetric.Type != "longitude" {
		return
	}

	for _, record := range records {
		timestamp := record.Timestamp.Unix() // Timestamp in seconds

		// Interpolate latitude and longitude
		latitude := interpolateLatitude(timestamp, latitudeMetric)
		longitude := interpolateLongitude(timestamp, longitudeMetric)

		// Use the library's methods to set latitude and longitude in degrees
		record.SetPositionLatDegrees(latitude)
		record.SetPositionLongDegrees(longitude)
	}
}

func convertStepsToCadence(stepsMetric types.Metric) types.Metric {
	cadenceMetric := types.Metric{
		Type:   "cadence",
		Unit:   "rpm",
		Values: []types.MetricValue{},
	}

	// Define the number of steps per revolution
	stepsPerRevolution := 2

	// Calculate RPM for each interval
	for _, interval := range stepsMetric.Values {
		metricValue := types.MetricValue{
			StartEpochMs: interval.StartEpochMs,
			EndEpochMs:   interval.EndEpochMs,
			Value:        0,
		}

		start := interval.StartEpochMs
		end := interval.EndEpochMs
		value := interval.Value

		// Calculate duration in minutes
		timeWindow := end - start
		durationMinutes := float64(timeWindow) / (1000 * 60)

		// Avoid division by zero
		if durationMinutes > 0 {
			stepsPerMinute := value / durationMinutes
			rpm := stepsPerMinute / float64(stepsPerRevolution)
			if rpm <= 180 {
				metricValue.Value = rpm
			}
			cadenceMetric.Values = append(cadenceMetric.Values, metricValue)
		}
	}

	return cadenceMetric
}

// fillCadenceFromSteps fills cadence for each record
// Interpolation is not really needed here
func fillCadenceFromSteps(records []*mesgdef.Record, stepsMetric types.Metric) {
	if stepsMetric.Type != "steps" || len(stepsMetric.Values) == 0 {
		return
	}

	// Skip if it's a default empty value (only one value with value 0)
	if len(stepsMetric.Values) == 1 && stepsMetric.Values[0].Value == 0 {
		return
	}

	cadence := convertStepsToCadence(stepsMetric)

	// Fill cadence for each record
	for _, record := range records {
		timestamp := record.Timestamp.Unix() // Timestamp in seconds

		for i, interval := range cadence.Values {
			// Convert start and end times to seconds
			currentStartSeconds := interval.StartEpochMs / 1000
			currentEndSeconds := interval.EndEpochMs / 1000

			// Special case: Handle the very first interval
			if i == 0 && timestamp >= currentStartSeconds && timestamp < currentEndSeconds {
				// Interpolate cadence within the first interval
				intervalDuration := float64(currentEndSeconds - currentStartSeconds)
				timeElapsed := float64(timestamp - currentStartSeconds)
				interpolatedCadence := interval.Value * (timeElapsed / intervalDuration)

				// Set the interpolated cadence value for the record
				record.Cadence = uint8(interpolatedCadence)
				break
			}

			// Handle timestamps within other intervals
			if timestamp >= currentStartSeconds && timestamp < currentEndSeconds {
				record.Cadence = uint8(interval.Value)
				break
			}
		}
	}
}

// fillDistance fills cumulated distance for each record
func fillDistance(records []*mesgdef.Record, distanceMetric types.Metric) {
	if distanceMetric.Type != "distance" {
		return
	}

	// Calculate cumulative distances
	cumulativeDistanceMetric := calculateCumulativeDistanceMetric(distanceMetric)
	first := cumulativeDistanceMetric.Values[0]
	last := cumulativeDistanceMetric.Values[len(cumulativeDistanceMetric.Values)-1]

	// Iterate through the records
	for _, record := range records {
		timestamp := record.Timestamp.Unix() // Timestamp in seconds

		// Handle timestamps before the first interval ends
		firstEndSeconds := first.EndEpochMs / 1000
		if timestamp < firstEndSeconds {
			// Calculate the proportion of time elapsed before the first interval ends
			totalIntervalDuration := float64(firstEndSeconds - records[0].Timestamp.Unix())
			timeElapsed := float64(timestamp - records[0].Timestamp.Unix())

			// Extrapolate the distance proportionally
			if totalIntervalDuration > 0 {
				extrapolatedDistanceMeters := (first.Value * 1000) * (timeElapsed / totalIntervalDuration)

				// Ensure the extrapolated distance is not negative
				if extrapolatedDistanceMeters < 0 {
					extrapolatedDistanceMeters = 0
				}

				record.SetDistanceScaled(extrapolatedDistanceMeters)
			} else {
				// If the interval duration is zero, set distance to 0
				record.SetDistanceScaled(0)
			}
			continue
		}

		// Iterate through the intervals
		for i := 0; i < len(cumulativeDistanceMetric.Values)-1; i++ {
			current := cumulativeDistanceMetric.Values[i]
			next := cumulativeDistanceMetric.Values[i+1]

			// Convert start and end times to seconds
			currentEndSeconds := current.EndEpochMs / 1000
			nextEndSeconds := next.EndEpochMs / 1000

			// Handle timestamps between intervals
			if timestamp >= currentEndSeconds && timestamp < nextEndSeconds {
				// Calculate interval duration and time elapsed in seconds
				intervalDuration := float64(nextEndSeconds - currentEndSeconds)
				timeElapsed := float64(timestamp - currentEndSeconds)

				distanceDeltaMeters := (next.Value - current.Value) * 1000 // Convert delta to meters
				interpolatedDistanceMeters := (current.Value * 1000) + (distanceDeltaMeters * (timeElapsed / intervalDuration))

				record.SetDistanceScaled(interpolatedDistanceMeters)
				break
			}
		}

		// Handle timestamps after the last interval
		lastEndSeconds := last.EndEpochMs / 1000
		if timestamp >= lastEndSeconds {
			// Set the last distance value using SetDistanceScaled (convert to meters)
			record.SetDistanceScaled(last.Value * 1000)
		}
	}
}

func calculateCumulativeDistanceMetric(distanceMetric types.Metric) types.Metric {
	cumulativeDistanceMetric := types.Metric{
		Type:   "cumulative_distance",
		Unit:   "KM",
		Values: []types.MetricValue{},
	}

	// Calculate cumulative distances
	cumulativeDistance := 0.0
	for _, value := range distanceMetric.Values {
		cumulativeDistance += value.Value
		cumulativeDistanceMetric.Values = append(cumulativeDistanceMetric.Values, types.MetricValue{
			StartEpochMs: value.StartEpochMs,
			EndEpochMs:   value.EndEpochMs,
			Value:        cumulativeDistance,
		})
	}

	return cumulativeDistanceMetric
}

// fillElevation fills altitude for each record
func fillElevation(records []*mesgdef.Record, elevationMetric types.Metric) {
	if elevationMetric.Type != "elevation" || len(elevationMetric.Values) == 0 {
		return
	}

	for _, record := range records {
		timestamp := record.Timestamp.Unix() // Timestamp in seconds

		// Handle timestamps before the first interval
		first := elevationMetric.Values[0]
		firstStartSeconds := first.StartEpochMs / 1000
		if timestamp < firstStartSeconds {
			// Set the altitude to the first interval's start value
			record.SetAltitudeScaled(first.Value)
			record.SetEnhancedAltitudeScaled(first.Value)
			continue
		}

		// Iterate through elevation intervals
		for i := 0; i < len(elevationMetric.Values)-1; i++ {
			current := elevationMetric.Values[i]
			next := elevationMetric.Values[i+1]

			// Convert start and end times to seconds
			currentStartSeconds := current.StartEpochMs / 1000
			nextStartSeconds := next.StartEpochMs / 1000

			// Check if the record's timestamp falls within the current interval
			if timestamp >= currentStartSeconds && timestamp < nextStartSeconds {
				// Calculate interval duration and time elapsed in seconds
				intervalDuration := float64(nextStartSeconds - currentStartSeconds)
				timeElapsed := float64(timestamp - currentStartSeconds)

				// Interpolate the elevation value
				elevationDelta := next.Value - current.Value
				interpolatedElevation := current.Value + (elevationDelta * (timeElapsed / intervalDuration))

				// Use the library's methods to set scaled altitude values
				record.SetAltitudeScaled(interpolatedElevation)
				record.SetEnhancedAltitudeScaled(interpolatedElevation)
				break
			}
		}

		// Handle timestamps after the last interval
		if len(elevationMetric.Values) > 0 {
			last := elevationMetric.Values[len(elevationMetric.Values)-1]
			if timestamp >= last.StartEpochMs/1000 {
				// Use the library's methods to set the last altitude value
				record.SetAltitudeScaled(last.Value)
				record.SetEnhancedAltitudeScaled(last.Value)
			}
		}
	}
}

// fillSpeedFromDistance calculates speed based on distance and time
func fillSpeedFromDistance(records []*mesgdef.Record) {
	for i := 1; i < len(records); i++ {
		// Get the current and previous records
		current := records[i]
		previous := records[i-1]

		// Calculate the time difference in seconds
		timeDelta := current.Timestamp.Sub(previous.Timestamp).Seconds()

		// Ensure timeDelta is greater than zero to avoid division by zero
		if timeDelta > 0 {
			// Calculate the distance difference in meters
			distanceDelta := current.DistanceScaled() - previous.DistanceScaled()

			// Compute the speed in meters per second
			speed := distanceDelta / timeDelta

			// Use the library's methods to set scaled speed values
			current.SetSpeedScaled(speed)
			current.SetEnhancedSpeedScaled(speed)
		} else {
			// If timeDelta is zero or negative, set speed to zero
			current.SetSpeedScaled(0)
			current.SetEnhancedSpeedScaled(0)
		}
	}
}

func (m *MetricsConverter) ParseSession(records []*mesgdef.Record) *mesgdef.Session {
	session := mesgdef.NewSession(nil)

	session.NumLaps = 1
	session.SetStartTime(utils.ParseTimeInMs(m.StartEpochMs))
	session.SetTimestamp(utils.ParseTimeInMs(m.EndEpochMs))
	session.SetTotalElapsedTime(uint32(m.EndEpochMs - m.StartEpochMs))

	if m.ActiveDurationMs > 0 {
		session.SetTotalTimerTime(uint32(m.ActiveDurationMs))
	}

	if m.DistanceSummary.Metric == "distance" {
		session.SetTotalDistanceScaled(m.DistanceSummary.Value * 1000)
	}

	if m.AscentSummary.Metric == "ascent" {
		session.SetTotalAscent(uint16(m.AscentSummary.Value))
	}

	if m.DescentSummary.Metric == "descent" {
		session.SetTotalDescent(uint16(m.DescentSummary.Value))
	}

	if m.CaloriesSummary.Metric == "calories" {
		session.SetTotalCalories(uint16(m.CaloriesSummary.Value))
	}

	if m.SpeedSummary.Metric == "speed" {
		session.SetAvgSpeedScaled(m.SpeedSummary.Value / 3.6)
	}

	if m.StepsSummary.Metric == "steps" && m.StepsSummary.Value > 0 {
		session.SetTotalCycles(uint32(m.StepsSummary.Value / 2))

		// Compute cadence based on total steps and active duration
		if m.ActiveDurationMs > 0 {
			cadence := computeCadenceInSpm(m.StepsSummary.Value, m.ActiveDurationMs)
			session.SetAvgCadence(uint8(cadence))
		}
	}

	// Set sport and subsport
	session.SetSport(typedef.SportRunning)
	if m.Indoor {
		session.SetSubSport(typedef.SubSportTreadmill)
	} else {
		session.SetSubSport(typedef.SubSportStreet)
	}

	// Compute max speed by looping over records
	maxSpeed := computeMaxSpeed(records)
	if maxSpeed > 0 {
		session.SetMaxSpeedScaled(maxSpeed)
	}

	return session
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

// computeCadenceInSpm calculates cadence in steps per minute (SPM)
func computeCadenceInSpm(totalSteps float64, activeDurationMs int64) float64 {
	// Convert active duration from milliseconds to minutes
	activeDurationMinutes := float64(activeDurationMs) / (1000 * 60)

	// Avoid division by zero
	if activeDurationMinutes > 0 {
		// Calculate steps per minute (SPM)
		stepsPerMinute := totalSteps / activeDurationMinutes

		return stepsPerMinute
	}

	// Return 0 if active duration is zero
	return 0
}

func computeMaxSpeed(records []*mesgdef.Record) float64 {
	maxSpeed := 0.0

	for _, record := range records {
		// Get the speed for the current record
		speed := record.SpeedScaled()

		// Update maxSpeed if the current speed is greater
		if speed > maxSpeed {
			maxSpeed = speed
		}
	}

	return maxSpeed
}
