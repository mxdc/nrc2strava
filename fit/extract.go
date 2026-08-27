package fit

import (
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
)

// extractSession finds and returns the session message from a FIT file
func extractSession(fit *proto.FIT) *mesgdef.Session {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumSession {
			return mesgdef.NewSession(&mesg)
		}
	}
	return nil
}

// extractFileId finds and returns the FileId message from a FIT file
func extractFileId(fit *proto.FIT) *mesgdef.FileId {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumFileId {
			return mesgdef.NewFileId(&mesg)
		}
	}
	return nil
}

// extractRecords finds and returns all record messages from a FIT file
func extractRecords(fit *proto.FIT) []*mesgdef.Record {
	var records []*mesgdef.Record
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumRecord {
			records = append(records, mesgdef.NewRecord(&mesg))
		}
	}
	return records
}

// extractEvents finds and returns all event messages from a FIT file
func extractEvents(fit *proto.FIT) []*mesgdef.Event {
	var events []*mesgdef.Event
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumEvent {
			events = append(events, mesgdef.NewEvent(&mesg))
		}
	}
	return events
}

// extractLaps finds and returns all lap messages from a FIT file
func extractLaps(fit *proto.FIT) []*mesgdef.Lap {
	var laps []*mesgdef.Lap
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumLap {
			laps = append(laps, mesgdef.NewLap(&mesg))
		}
	}
	return laps
}

// extractActivity finds and returns the activity message from a FIT file
func extractActivity(fit *proto.FIT) *mesgdef.Activity {
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumActivity {
			return mesgdef.NewActivity(&mesg)
		}
	}
	return nil
}

// extractDeveloperDataIds finds and returns all DeveloperDataId messages from a FIT file
func extractDeveloperDataIds(fit *proto.FIT) []*mesgdef.DeveloperDataId {
	var ids []*mesgdef.DeveloperDataId
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumDeveloperDataId {
			ids = append(ids, mesgdef.NewDeveloperDataId(&mesg))
		}
	}
	return ids
}

// extractFieldDescriptions finds and returns all FieldDescription messages from a FIT file
func extractFieldDescriptions(fit *proto.FIT) []*mesgdef.FieldDescription {
	var descriptions []*mesgdef.FieldDescription
	for _, mesg := range fit.Messages {
		if mesg.Num == typedef.MesgNumFieldDescription {
			descriptions = append(descriptions, mesgdef.NewFieldDescription(&mesg))
		}
	}
	return descriptions
}
