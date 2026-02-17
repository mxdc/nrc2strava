package converter

import (
	"log"
	"os"

	"github.com/muktihari/fit/profile"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
	"github.com/mxdc/nrc2strava/types"
	"github.com/mxdc/nrc2strava/utils"
)

// ActivitiesConverter converts the activities into the FIT Activity format
type ActivitiesConverter struct {
	// logger
	logger *log.Logger
}

// InitActivitiesConverter returns an initialized ActivitiesConverter
func InitActivitiesConverter() *ActivitiesConverter {
	var parser ActivitiesConverter

	parser.logger = log.New(os.Stderr, "", log.LstdFlags)

	return &parser
}

// LoadActivities load JSON files into memory
func (c *ActivitiesConverter) ConvertRun(nikeActivity *types.Activity) types.Run {
	activity := filedef.NewActivity()

	// FileId
	activity.FileId = *mesgdef.NewFileId(nil).
		SetType(typedef.FileActivity).
		SetTimeCreated(utils.ParseTimeInMs(nikeActivity.StartEpochMs)).
		SetManufacturer(typedef.ManufacturerNike).
		SetSerialNumber(12345)

	// Activity Title
	activityTitle := getRunName(nikeActivity.Tags, nikeActivity.StartEpochMs)
	fieldBase := &proto.FieldBase{
		Num:        99,
		Name:       "Title",
		Type:       profile.String,
		BaseType:   basetype.String,
		Array:      false,
		Accumulate: false,
		Scale:      1,
		Offset:     0,
		Units:      "",
	}

	customField := proto.Field{FieldBase: fieldBase, Value: proto.String(activityTitle)}
	activity.FileId.SetUnknownFields(customField)

	// DeveloperDataIds
	activity.DeveloperDataIds = append(
		activity.DeveloperDataIds,
		mesgdef.NewDeveloperDataId(nil).
			SetApplicationId([]byte{99}).
			SetDeveloperDataIndex(0),
	)

	// events
	events := convertMomentsToEvents(nikeActivity)
	activity.Events = events

	// records
	metricsConverter := InitMetricsConverter(
		nikeActivity.StartEpochMs,
		nikeActivity.EndEpochMs,
		nikeActivity.ActiveDuration,
		nikeActivity.Metrics,
		nikeActivity.Summaries,
		nikeActivity.Moments,
		nikeActivity.Tags,
	)

	records := metricsConverter.ParseRecords()
	// printRecordLines(records)
	activity.Records = records

	// lap
	activity.Laps = append(
		activity.Laps,
		mesgdef.NewLap(nil).
			SetStartTime(utils.ParseTimeInMs(nikeActivity.StartEpochMs)).
			SetTimestamp(utils.ParseTimeInMs(nikeActivity.EndEpochMs)).
			SetTotalElapsedTime(uint32(nikeActivity.EndEpochMs-nikeActivity.StartEpochMs)).
			SetTotalTimerTime(uint32(nikeActivity.ActiveDuration)),
	)

	// session
	session := metricsConverter.ParseSession(records)
	activity.Sessions = append(
		activity.Sessions,
		session,
	)

	activity.Activity = mesgdef.NewActivity(nil).
		SetType(typedef.Activity(typedef.ActivityTypeRunning)).
		SetTimestamp(utils.ParseTimeInMs(nikeActivity.EndEpochMs)).
		SetNumSessions(1)

	return types.Run{
		Id:       nikeActivity.ID,
		Activity: activity,
	}
}

func convertMomentsToEvents(nikeActivity *types.Activity) []*mesgdef.Event {
	events := []*mesgdef.Event{}

	// Add start
	events = append(events, mesgdef.NewEvent(nil).
		SetTimestamp(utils.ParseTimeInMs(nikeActivity.StartEpochMs)).
		SetEventType(typedef.EventTypeStart).
		SetEvent(typedef.EventTimer),
	)

	// Add pauses
	for _, moment := range nikeActivity.Moments {
		if moment.Key == "halt" {
			events = append(events, mesgdef.NewEvent(nil).
				SetTimestamp(utils.ParseTimeInMs(moment.Timestamp)).
				SetEventType(convertMomentEventType(moment.Value)).
				SetEvent(typedef.EventTimer),
			)
		}
	}

	// Add finish
	events = append(events, mesgdef.NewEvent(nil).
		SetTimestamp(utils.ParseTimeInMs(nikeActivity.EndEpochMs)).
		SetEventType(typedef.EventTypeStopAll).
		SetEvent(typedef.EventTimer),
	)

	return events
}

func convertMomentEventType(value string) typedef.EventType {
	if value == "pause" {
		return typedef.EventTypeStop
	}

	if value == "resume" {
		return typedef.EventTypeStart
	}

	return typedef.EventTypeInvalid
}

func getRunName(tags map[string]string, StartEpochMs int64) string {
	if name, ok := tags["com.nike.name"]; ok {
		return name
	}

	parsedTime := utils.ParseTimeInMs(StartEpochMs)
	// Format as "YYYY-MM-DD"
	date := parsedTime.Format("2006-01-02")
	// Format as "HH:mm"
	time := parsedTime.Format("15h04")

	return "Run on " + date + " at " + time
}
