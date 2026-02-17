package nrc

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// NikeDownloader represents the Nike API client
type NikeDownloader struct {
	downloadActivitiesDir string
	nikeApi               *NikeApi
	logger                *logrus.Logger
}

// NewNikeDownloader initializes a new NewNikeDownloader instance
func NewNikeDownloader(nikeApi *NikeApi, downloadActivitiesDir string) *NikeDownloader {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &NikeDownloader{
		downloadActivitiesDir: downloadActivitiesDir,
		nikeApi:               nikeApi,
		logger:                logger,
	}
}

func (n *NikeDownloader) DownloadActivities() {
	n.logger.Info("Downloading activities")

	// Create the directory if it doesn't exist
	if _, err := os.Stat(n.downloadActivitiesDir); os.IsNotExist(err) {
		if err := os.Mkdir(n.downloadActivitiesDir, os.ModePerm); err != nil {
			n.logger.Errorf("Error creating activities folder: %v\n", err)
			return
		}
	}

	activities, err := n.nikeApi.GetActivityList()
	if err != nil {
		n.logger.Errorf("Error fetching activity list: %v\n", err)
		return
	}

	total := len(activities)
	n.logger.Infof("Total activity(s) to download: %d\n", total)

	for index, activityID := range activities {
		n.logger.Infof("Downloading activity ID: %s\n", activityID)

		activityDetails, err := n.nikeApi.GetActivityDetails(activityID)
		if err != nil {
			n.logger.Errorf("Error downloading activity ID %s: %v\n", activityID, err)
			continue
		}

		filepath := filepath.Join(n.downloadActivitiesDir, fmt.Sprintf("%s.json", activityID))
		err = n.SaveActivity(activityDetails, filepath)
		if err != nil {
			n.logger.Errorf("Error saving activity ID %s: %v\n", activityID, err)
			continue
		}

		if index < total-1 {
			n.logger.Debug("Waiting for 200ms before downloading the next activity...")
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (n *NikeDownloader) SaveActivity(activityDetails []byte, filepath string) error {
	n.logger.Debugf("Storing activity: %s\n", filepath)

	if err := os.WriteFile(filepath, activityDetails, 0644); err != nil {
		return err
	}

	n.logger.Debugf("Activity stored successfully to %s\n", filepath)
	return nil
}
