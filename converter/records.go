package converter

import (
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/mxdc/nrc2strava/types"
)

func interpolateValue(timestamp int64, metric types.Metric) float64 {
	// Handle timestamps before the first interval
	if len(metric.Values) > 0 {
		first := metric.Values[0]
		firstStartSeconds := first.StartEpochMs / 1000
		if timestamp < firstStartSeconds {
			return first.Value
		}
	}

	// Iterate through intervals
	for i := 0; i < len(metric.Values)-1; i++ {
		current := metric.Values[i]
		next := metric.Values[i+1]

		// Convert start and end times to seconds
		currentStartSeconds := current.StartEpochMs / 1000
		nextStartSeconds := next.StartEpochMs / 1000

		// Check if the timestamp falls within the current interval
		if timestamp >= currentStartSeconds && timestamp < nextStartSeconds {
			intervalDuration := float64(nextStartSeconds - currentStartSeconds)
			timeElapsed := float64(timestamp - currentStartSeconds)
			delta := next.Value - current.Value
			return current.Value + (delta * (timeElapsed / intervalDuration))
		}
	}

	// Handle timestamps after the last interval
	last := metric.Values[len(metric.Values)-1]
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
		latitude := interpolateValue(timestamp, latitudeMetric)
		longitude := interpolateValue(timestamp, longitudeMetric)

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
