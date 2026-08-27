package fit

import (
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/proto"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// RecordDeduplicator merges FIT records that share a timestamp into one.
type RecordDeduplicator struct {
	// logger
	logger *logrus.Logger
}

// NewRecordDeduplicator returns an initialized RecordDeduplicator
func NewRecordDeduplicator() *RecordDeduplicator {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &RecordDeduplicator{
		logger: logger,
	}
}

// Deduplicate merges records with the same timestamp into single records
func (d *RecordDeduplicator) Deduplicate(records []*mesgdef.Record) []*mesgdef.Record {
	if len(records) == 0 {
		return records
	}

	// Group records by timestamp, filtering out empty records
	recordsByTime := make(map[time.Time][]*mesgdef.Record)
	emptyRecordCount := 0
	for _, record := range records {
		if isEmptyRecord(record) {
			emptyRecordCount++
			continue
		}
		recordsByTime[record.Timestamp] = append(recordsByTime[record.Timestamp], record)
	}

	// Merge records with the same timestamp
	deduplicated := make([]*mesgdef.Record, 0, len(recordsByTime))
	processedTimes := make(map[time.Time]bool)

	// Iterate through original records to preserve order
	for _, record := range records {
		if isEmptyRecord(record) || processedTimes[record.Timestamp] {
			continue
		}
		processedTimes[record.Timestamp] = true

		duplicates := recordsByTime[record.Timestamp]
		if len(duplicates) == 1 {
			// No duplicates, use as-is
			deduplicated = append(deduplicated, duplicates[0])
		} else {
			// Merge all duplicate records into one
			mergedRecord := mergeRecordFields(duplicates)
			deduplicated = append(deduplicated, mergedRecord)
		}
	}

	originalCount := len(records)
	deduplicatedCount := len(deduplicated)
	if originalCount != deduplicatedCount {
		duplicatesRemoved := originalCount - deduplicatedCount - emptyRecordCount
		d.logger.Infof("Deduplicated records: %d → %d (removed %d empty, %d duplicates)",
			originalCount, deduplicatedCount, emptyRecordCount, duplicatesRemoved)
	} else if emptyRecordCount > 0 {
		d.logger.Infof("Removed %d empty records", emptyRecordCount)
	}

	return deduplicated
}

// isEmptyRecord checks if a record has no meaningful data (no GPS, speed, altitude, etc.)
func isEmptyRecord(record *mesgdef.Record) bool {
	// A record is considered empty if it has no GPS coordinates and no other useful data
	hasData := false

	// Check if it has valid GPS coordinates
	if record.PositionLat != 0x7FFFFFFF && record.PositionLong != 0x7FFFFFFF {
		hasData = true
	}

	// Check if it has valid speed
	if record.Speed != 0xFFFF && record.Speed > 0 {
		hasData = true
	}

	// Check if it has valid altitude
	if record.Altitude != 0xFFFF && record.EnhancedAltitude != 0xFFFF {
		hasData = true
	}

	// Check if it has valid heart rate
	if record.HeartRate != 0xFF && record.HeartRate > 0 {
		hasData = true
	}

	// Check if it has valid cadence
	if record.Cadence != 0xFF && record.Cadence > 0 {
		hasData = true
	}

	return !hasData
}

// mergeRecordFields merges multiple records with the same timestamp into one record
func mergeRecordFields(records []*mesgdef.Record) *mesgdef.Record {
	if len(records) == 0 {
		return nil
	}
	if len(records) == 1 {
		return records[0]
	}

	// Start with the first record's message
	baseMesg := records[0].ToMesg(nil)

	// Create a map of existing fields by field number
	fieldsByNum := make(map[byte]proto.Field)
	for _, field := range baseMesg.Fields {
		fieldsByNum[field.Num] = field
	}

	// Merge fields from other records
	for i := 1; i < len(records); i++ {
		mesg := records[i].ToMesg(nil)
		for _, field := range mesg.Fields {
			// Only add field if it doesn't exist or if the existing field is empty/invalid
			existing, exists := fieldsByNum[field.Num]
			if !exists {
				fieldsByNum[field.Num] = field
			} else {
				// Replace if existing value is invalid/empty and new value is not
				existingValue := existing.Value.Any()
				newValue := field.Value.Any()
				if isEmptyValue(existingValue) && !isEmptyValue(newValue) {
					fieldsByNum[field.Num] = field
				}
			}
		}

		// Merge DeveloperFields
		for _, devField := range mesg.DeveloperFields {
			baseMesg.DeveloperFields = append(baseMesg.DeveloperFields, devField)
		}
	}

	// Rebuild fields slice from map
	baseMesg.Fields = make([]proto.Field, 0, len(fieldsByNum))
	for _, field := range fieldsByNum {
		baseMesg.Fields = append(baseMesg.Fields, field)
	}

	return mesgdef.NewRecord(&baseMesg)
}

// isEmptyValue checks if a field value is empty/invalid
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case uint8:
		return v == 0xFF
	case uint16:
		return v == 0xFFFF
	case uint32:
		return v == 0xFFFFFFFF
	case int8:
		return v == 0x7F
	case int16:
		return v == 0x7FFF
	case int32:
		return v == 0x7FFFFFFF
	case float32:
		return v == 0.0
	case float64:
		return v == 0.0
	case string:
		return v == ""
	case time.Time:
		return v.IsZero()
	default:
		return false
	}
}
