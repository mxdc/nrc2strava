package migrator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mxdc/nrc2strava/converter"
	"github.com/mxdc/nrc2strava/fit"
	"github.com/mxdc/nrc2strava/nrc"
	"github.com/mxdc/nrc2strava/strava"
	"github.com/mxdc/nrc2strava/types"
)

// Migrator represents the Migrator client
type Migrator struct {
	nikeApi      *nrc.NikeApi
	stravaWeb    *strava.StravaWeb
	FitOutputDir string
	logger       *log.Logger
}

// NewMigrator initializes a new NewMigrator instance
func NewMigrator(nikeApi *nrc.NikeApi, stravaWeb *strava.StravaWeb, FitOutputDir string) *Migrator {
	return &Migrator{
		nikeApi:      nikeApi,
		stravaWeb:    stravaWeb,
		FitOutputDir: FitOutputDir,
		logger:       log.New(os.Stderr, "", log.LstdFlags),
	}
}

// MigrateActivities migrates activities from Nike to Strava
func (m *Migrator) MigrateActivities() {
	activitiesIds, err := m.nikeApi.GetActivityList()
	if err != nil {
		m.logger.Printf("Error fetching activity list: %v\n", err)
		return
	}

	m.logger.Printf("Total activity(s) to migrate: %d\n", len(activitiesIds))

	activitiesConverter := converter.InitActivitiesConverter()
	activityWriter := fit.InitActivityWriter(m.FitOutputDir)

	total := len(activitiesIds)
	for index, activityID := range activitiesIds {
		m.logger.Printf("Migrating activity ID: %s\n", activityID)

		// Fetch activity details with retry logic
		activityDetails, err := m.nikeApi.GetActivityDetailsWithRetry(activityID, 3)
		if err != nil {
			m.logger.Printf("Migration error: %v\n", err)
			continue
		}

		// Unmarshal JSON into Go struct
		var activity types.Activity
		err = json.Unmarshal(activityDetails, &activity)
		if err != nil {
			m.logger.Printf("Error parsing JSON:, %v", err)
			continue
		}

		run := activitiesConverter.ConvertRun(&activity)
		outputFilename := activityWriter.WriteFIT(run)
		stravaUploader := strava.NewStravaUploader(outputFilename, m.stravaWeb)
		stravaUploader.UploadActivity(outputFilename)
		if index < total-1 {
			fmt.Println("Waiting for 10 seconds before processing the next file...")
			time.Sleep(10 * time.Second)
		}
	}
}
