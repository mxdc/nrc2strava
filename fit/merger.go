package fit

import (
	"fmt"
	"os"
	"time"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// ActivityMerger merges two FIT activities into one
type ActivityMerger struct {
	FirstActivityPath  string
	SecondActivityPath string
	OutputPath         string

	firstFit  *proto.FIT
	secondFit *proto.FIT

	logger *logrus.Logger
}

// NewActivityMerger creates a new ActivityMerger instance
func NewActivityMerger(firstPath, secondPath, outputPath string) *ActivityMerger {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &ActivityMerger{
		FirstActivityPath:  firstPath,
		SecondActivityPath: secondPath,
		OutputPath:         outputPath,
		logger:             logger,
	}
}

// MergeActivities performs the complete merge operation
func (m *ActivityMerger) MergeActivities() error {
	m.logger.Info("Starting merge process...")

	// Phase 1: Read and validate
	if err := m.readFitFiles(); err != nil {
		return fmt.Errorf("error reading FIT files: %w", err)
	}

	if err := m.validateCompatibility(); err != nil {
		return fmt.Errorf("activities are not compatible: %w", err)
	}

	// Phase 2: Merge
	mergedFit, err := m.mergeData()
	if err != nil {
		return fmt.Errorf("error merging data: %w", err)
	}

	// Phase 3: Write output
	if err := m.writeMergedFile(mergedFit); err != nil {
		return fmt.Errorf("error writing merged file: %w", err)
	}

	m.logger.Infof("✓ Successfully merged activities to: %s", m.OutputPath)
	return nil
}

// readFitFiles reads and decodes both FIT files
func (m *ActivityMerger) readFitFiles() error {
	m.logger.Info("Reading FIT files...")

	// Read first file
	file1, err := os.Open(m.FirstActivityPath)
	if err != nil {
		return fmt.Errorf("failed to open first file: %w", err)
	}
	defer file1.Close()

	dec1 := decoder.New(file1)
	fit1, err := dec1.Decode()
	if err != nil {
		return fmt.Errorf("failed to decode first FIT file: %w", err)
	}
	m.firstFit = fit1
	m.logger.Infof("✓ Read first activity: %s", m.FirstActivityPath)

	// Read second file
	file2, err := os.Open(m.SecondActivityPath)
	if err != nil {
		return fmt.Errorf("failed to open second file: %w", err)
	}
	defer file2.Close()

	dec2 := decoder.New(file2)
	fit2, err := dec2.Decode()
	if err != nil {
		return fmt.Errorf("failed to decode second FIT file: %w", err)
	}
	m.secondFit = fit2
	m.logger.Infof("✓ Read second activity: %s", m.SecondActivityPath)

	return nil
}

// validateCompatibility checks if the two activities can be merged
func (m *ActivityMerger) validateCompatibility() error {
	m.logger.Info("Validating activity compatibility...")

	session1 := m.extractSession(m.firstFit)
	session2 := m.extractSession(m.secondFit)

	if session1 == nil {
		return fmt.Errorf("first activity has no session data")
	}
	if session2 == nil {
		return fmt.Errorf("second activity has no session data")
	}

	// Check sport type
	if session1.Sport != session2.Sport {
		return fmt.Errorf("sport type mismatch: %v vs %v", session1.Sport, session2.Sport)
	}

	// Check sub-sport
	if session1.SubSport != session2.SubSport {
		return fmt.Errorf("sub-sport mismatch: %v vs %v (both must be same type - e.g., treadmill or outdoor)",
			session1.SubSport, session2.SubSport)
	}

	m.logger.Infof("✓ Activities compatible: Sport=%v, SubSport=%v", session1.Sport, session1.SubSport)
	m.logActivityDetails("Activity 1", session1)
	m.logActivityDetails("Activity 2", session2)

	return nil
}

// extractSession finds and returns the session message from a FIT file
func (m *ActivityMerger) extractSession(fit *proto.FIT) *mesgdef.Session {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumSession {
			return mesgdef.NewSession(&mesg)
		}
	}
	return nil
}

// extractFileId finds and returns the FileId message from a FIT file
func (m *ActivityMerger) extractFileId(fit *proto.FIT) *mesgdef.FileId {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumFileId {
			return mesgdef.NewFileId(&mesg)
		}
	}
	return nil
}

// extractRecords finds and returns all record messages from a FIT file
func (m *ActivityMerger) extractRecords(fit *proto.FIT) []*mesgdef.Record {
	var records []*mesgdef.Record
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumRecord {
			records = append(records, mesgdef.NewRecord(&mesg))
		}
	}
	return records
}

// extractEvents finds and returns all event messages from a FIT file
func (m *ActivityMerger) extractEvents(fit *proto.FIT) []*mesgdef.Event {
	var events []*mesgdef.Event
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumEvent {
			events = append(events, mesgdef.NewEvent(&mesg))
		}
	}
	return events
}

// extractLaps finds and returns all lap messages from a FIT file
func (m *ActivityMerger) extractLaps(fit *proto.FIT) []*mesgdef.Lap {
	var laps []*mesgdef.Lap
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumLap {
			laps = append(laps, mesgdef.NewLap(&mesg))
		}
	}
	return laps
}

// extractActivity finds and returns the activity message from a FIT file
func (m *ActivityMerger) extractActivity(fit *proto.FIT) *mesgdef.Activity {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumActivity {
			return mesgdef.NewActivity(&mesg)
		}
	}
	return nil
}

// extractDeveloperDataIds finds and returns all DeveloperDataId messages from a FIT file
func (m *ActivityMerger) extractDeveloperDataIds(fit *proto.FIT) []*mesgdef.DeveloperDataId {
	var ids []*mesgdef.DeveloperDataId
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumDeveloperDataId {
			ids = append(ids, mesgdef.NewDeveloperDataId(&mesg))
		}
	}
	return ids
}

// extractFieldDescriptions finds and returns all FieldDescription messages from a FIT file
func (m *ActivityMerger) extractFieldDescriptions(fit *proto.FIT) []*mesgdef.FieldDescription {
	var descriptions []*mesgdef.FieldDescription
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumFieldDescription {
			descriptions = append(descriptions, mesgdef.NewFieldDescription(&mesg))
		}
	}
	return descriptions
}

// logActivityDetails logs details about an activity session
func (m *ActivityMerger) logActivityDetails(label string, session *mesgdef.Session) {
	distance := float64(session.TotalDistance) / 1000.0 // Convert to km
	duration := time.Duration(session.TotalTimerTime) * time.Millisecond

	m.logger.Infof("  %s: Distance=%.2f km, Duration=%v", label, distance, duration)
}

// mergeData performs the actual merging of the two activities
func (m *ActivityMerger) mergeData() (*proto.FIT, error) {
	m.logger.Info("Merging activity data...")

	// Extract data from both activities
	session1 := m.extractSession(m.firstFit)
	session2 := m.extractSession(m.secondFit)
	records1 := m.extractRecords(m.firstFit)
	records2 := m.extractRecords(m.secondFit)
	events1 := m.extractEvents(m.firstFit)
	events2 := m.extractEvents(m.secondFit)
	laps1 := m.extractLaps(m.firstFit)
	laps2 := m.extractLaps(m.secondFit)
	fileId1 := m.extractFileId(m.firstFit)
	activity1 := m.extractActivity(m.firstFit)
	developerDataIds1 := m.extractDeveloperDataIds(m.firstFit)
	fieldDescriptions1 := m.extractFieldDescriptions(m.firstFit)

	if len(records1) == 0 {
		return nil, fmt.Errorf("first activity has no records")
	}
	if len(records2) == 0 {
		return nil, fmt.Errorf("second activity has no records")
	}

	// Calculate offsets
	lastRecord1 := records1[len(records1)-1]
	firstRecord2 := records2[0]

	timeOffset := lastRecord1.Timestamp.Sub(firstRecord2.Timestamp)
	m.logger.Infof("Time offset: %v", timeOffset)

	distanceOffset := float64(lastRecord1.Distance)
	m.logger.Infof("Distance offset: %.2f meters", distanceOffset)

	// Merge records
	mergedRecords := m.mergeRecords(records1, records2, timeOffset, distanceOffset)
	m.logger.Infof("✓ Merged %d records (Activity1: %d, Activity2: %d)",
		len(mergedRecords), len(records1), len(records2))

	// Merge events
	mergedEvents := m.mergeEvents(events1, events2, timeOffset)
	m.logger.Infof("✓ Merged %d events", len(mergedEvents))

	// Merge laps into a single lap
	mergedLap := m.createMergedLap(laps1, laps2, session1, session2, mergedRecords, timeOffset)
	m.logger.Info("✓ Created merged lap")

	// Create merged session
	mergedSession := m.createMergedSession(session1, session2, mergedRecords, timeOffset)
	m.logger.Info("✓ Created merged session")

	// Create merged activity message
	mergedActivity := m.createMergedActivity(activity1, mergedSession, mergedRecords)

	// Build the merged FIT file
	mergedFit := &proto.FIT{
		Messages: []proto.Message{},
	}

	// Add FileId (from first activity)
	if fileId1 != nil {
		mergedFit.Messages = append(mergedFit.Messages, fileId1.ToMesg(nil))
	}

	// Add DeveloperDataIds (from first activity)
	for _, devDataId := range developerDataIds1 {
		mergedFit.Messages = append(mergedFit.Messages, devDataId.ToMesg(nil))
	}

	// Add FieldDescriptions (from first activity)
	for _, fieldDesc := range fieldDescriptions1 {
		mergedFit.Messages = append(mergedFit.Messages, fieldDesc.ToMesg(nil))
	}

	// Add events
	for _, event := range mergedEvents {
		mergedFit.Messages = append(mergedFit.Messages, event.ToMesg(nil))
	}

	// Add records
	for _, record := range mergedRecords {
		mergedFit.Messages = append(mergedFit.Messages, record.ToMesg(nil))
	}

	// Add merged lap
	mergedFit.Messages = append(mergedFit.Messages, mergedLap.ToMesg(nil))

	// Add session
	mergedFit.Messages = append(mergedFit.Messages, mergedSession.ToMesg(nil))

	// Add activity
	mergedFit.Messages = append(mergedFit.Messages, mergedActivity.ToMesg(nil))

	return mergedFit, nil
}

// mergeRecords merges records from both activities
func (m *ActivityMerger) mergeRecords(records1, records2 []*mesgdef.Record, timeOffset time.Duration, distanceOffset float64) []*mesgdef.Record {
	merged := make([]*mesgdef.Record, 0, len(records1)+len(records2))

	// Add all records from first activity
	merged = append(merged, records1...)

	// Add records from second activity with adjustments
	for _, record := range records2 {
		// Convert to proto.Message to preserve all fields including UnknownFields and DeveloperFields
		mesg := record.ToMesg(nil)

		// Adjust timestamp and distance fields
		for i := range mesg.Fields {
			if mesg.Fields[i].Num == 253 { // 253 is the field number for timestamp
				if timestamp, ok := mesg.Fields[i].Value.Any().(time.Time); ok {
					mesg.Fields[i].Value = proto.Any(timestamp.Add(timeOffset))
				}
			} else if mesg.Fields[i].Num == 5 { // 5 is the field number for distance
				if dist, ok := mesg.Fields[i].Value.Any().(uint32); ok {
					newDist := float64(dist) + distanceOffset
					mesg.Fields[i].Value = proto.Uint32(uint32(newDist))
				}
			}
		}

		// Create new Record from modified message
		adjustedRecord := mesgdef.NewRecord(&mesg)
		merged = append(merged, adjustedRecord)
	}

	return merged
}

// mergeEvents merges events from both activities
func (m *ActivityMerger) mergeEvents(events1, events2 []*mesgdef.Event, timeOffset time.Duration) []*mesgdef.Event {
	merged := make([]*mesgdef.Event, 0, len(events1)+len(events2))

	// Add all events from first activity (no filtering)
	merged = append(merged, events1...)

	// Add all events from second activity with adjusted timestamps (no filtering)
	for _, event := range events2 {
		// Convert to proto.Message to preserve all fields including UnknownFields
		mesg := event.ToMesg(nil)

		// Adjust the timestamp field
		for i := range mesg.Fields {
			if mesg.Fields[i].Num == 253 { // 253 is the field number for timestamp
				if timestamp, ok := mesg.Fields[i].Value.Any().(time.Time); ok {
					mesg.Fields[i].Value = proto.Any(timestamp.Add(timeOffset))
				}
			}
		}

		// Create new Event from modified message
		adjustedEvent := mesgdef.NewEvent(&mesg)
		merged = append(merged, adjustedEvent)
	}

	return merged
}

// createMergedLap creates a single merged lap from both activities' laps
func (m *ActivityMerger) createMergedLap(laps1, laps2 []*mesgdef.Lap, session1, session2 *mesgdef.Session, mergedRecords []*mesgdef.Record, timeOffset time.Duration) *mesgdef.Lap {
	lap := mesgdef.NewLap(nil)

	// Use start time from first activity
	if len(laps1) > 0 {
		lap.SetStartTime(laps1[0].StartTime)
	} else {
		lap.SetStartTime(session1.StartTime)
	}

	// Use end time from last record
	lastRecord := mergedRecords[len(mergedRecords)-1]
	lap.SetTimestamp(lastRecord.Timestamp)

	// Calculate total elapsed time (sum of both sessions)
	totalElapsed := uint32(session1.TotalElapsedTime + session2.TotalElapsedTime)
	lap.SetTotalElapsedTime(totalElapsed)

	// Calculate total timer time (sum of both sessions)
	totalTimer := uint32(session1.TotalTimerTime + session2.TotalTimerTime)
	lap.SetTotalTimerTime(totalTimer)

	// Calculate total distance (sum of both sessions)
	totalDistance := float64(session1.TotalDistance + session2.TotalDistance)
	if totalDistance > 0 {
		lap.SetTotalDistanceScaled(totalDistance / 100.0)
	}

	return lap
}

// createMergedSession creates a single merged session from both activities
func (m *ActivityMerger) createMergedSession(session1, session2 *mesgdef.Session, mergedRecords []*mesgdef.Record, timeOffset time.Duration) *mesgdef.Session {
	session := mesgdef.NewSession(nil)

	// Set basic properties from first session
	session.SetSport(session1.Sport)
	session.SetSubSport(session1.SubSport)

	// Set start time from first activity
	session.SetStartTime(session1.StartTime)

	// Calculate total elapsed time
	totalElapsed := uint32(session1.TotalElapsedTime + session2.TotalElapsedTime)
	session.SetTotalElapsedTime(totalElapsed)

	// Calculate total timer time
	totalTimer := uint32(session1.TotalTimerTime + session2.TotalTimerTime)
	session.SetTotalTimerTime(totalTimer)

	// Calculate total distance
	totalDistance := float64(session1.TotalDistance + session2.TotalDistance)
	if totalDistance > 0 {
		session.SetTotalDistanceScaled(totalDistance / 100.0)
	}

	// Calculate total ascent
	totalAscent := session1.TotalAscent + session2.TotalAscent
	if totalAscent > 0 {
		session.SetTotalAscent(totalAscent)
	}

	// Calculate total descent
	totalDescent := session1.TotalDescent + session2.TotalDescent
	if totalDescent > 0 {
		session.SetTotalDescent(totalDescent)
	}

	// Calculate total cycles (steps/2)
	totalCycles := session1.TotalCycles + session2.TotalCycles
	if totalCycles > 0 {
		session.SetTotalCycles(totalCycles)
	}

	// Calculate max speed from records
	maxSpeed := m.calculateMaxSpeed(mergedRecords)
	if maxSpeed > 0 {
		session.SetMaxSpeedScaled(maxSpeed)
	}

	// Calculate weighted average cadence
	if totalTimer > 0 {
		totalWeightedCadence := float64(0)
		if session1.AvgCadence > 0 && session1.TotalTimerTime > 0 {
			totalWeightedCadence += float64(session1.AvgCadence) * float64(session1.TotalTimerTime)
		}
		if session2.AvgCadence > 0 && session2.TotalTimerTime > 0 {
			totalWeightedCadence += float64(session2.AvgCadence) * float64(session2.TotalTimerTime)
		}
		avgCadence := uint8(totalWeightedCadence / float64(totalTimer))
		if avgCadence > 0 {
			session.SetAvgCadence(avgCadence)
		}
	}

	// Set number of laps to 1 (we merged them into a single lap)
	session.NumLaps = 1

	return session
}

// calculateMaxSpeed finds the maximum speed from all records
func (m *ActivityMerger) calculateMaxSpeed(records []*mesgdef.Record) float64 {
	maxSpeed := 0.0
	for _, record := range records {
		if record.Speed > 0 {
			speed := float64(record.Speed) / 1000.0
			if speed > maxSpeed {
				maxSpeed = speed
			}
		}
	}
	return maxSpeed
}

// createMergedActivity creates the activity message
func (m *ActivityMerger) createMergedActivity(originalActivity *mesgdef.Activity, session *mesgdef.Session, records []*mesgdef.Record) *mesgdef.Activity {
	activity := mesgdef.NewActivity(nil)

	lastRecord := records[len(records)-1]
	activity.SetTimestamp(lastRecord.Timestamp)

	// Use activity type from original first activity
	if originalActivity != nil && originalActivity.Type != 0 {
		activity.SetType(originalActivity.Type)
	} else {
		activity.SetType(typedef.ActivityManual)
	}

	activity.SetNumSessions(1)

	if session.TotalTimerTime > 0 {
		activity.SetTotalTimerTime(session.TotalTimerTime)
	}

	// Add LocalTimestamp if present in original
	if originalActivity != nil && !originalActivity.LocalTimestamp.IsZero() {
		// Calculate the new local timestamp (add the time difference)
		timeDiff := lastRecord.Timestamp.Sub(originalActivity.Timestamp)
		newLocalTimestamp := originalActivity.LocalTimestamp.Add(timeDiff)
		activity.SetLocalTimestamp(newLocalTimestamp)
	}

	return activity
}

// writeMergedFile writes the merged FIT data to a file
func (m *ActivityMerger) writeMergedFile(fit *proto.FIT) error {
	m.logger.Infof("Writing merged file to: %s", m.OutputPath)

	f, err := os.OpenFile(m.OutputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer f.Close()

	enc := encoder.New(f)
	if err := enc.Encode(fit); err != nil {
		return fmt.Errorf("error encoding FIT file: %w", err)
	}

	return nil
}
