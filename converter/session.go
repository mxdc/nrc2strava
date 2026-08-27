package converter

import (
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/mxdc/nrc2strava/utils"
)

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
