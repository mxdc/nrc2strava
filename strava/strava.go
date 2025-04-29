package strava

import (
	"time"

	"github.com/sirupsen/logrus"
)

// StravaUploader represents the Strava API client
type StravaUploader struct {
	FitActivityFile string
	Client          *StravaWeb

	// logger
	logger *logrus.Logger
}

// NewStravaUploader initializes a new StravaUploader instance
func NewStravaUploader(fitActivityFile string, stravaWeb *StravaWeb) *StravaUploader {
	return &StravaUploader{
		FitActivityFile: fitActivityFile,
		Client:          stravaWeb,
		logger:          logrus.New(),
	}
}

func (s *StravaUploader) UploadActivity(fitActivityFilepath string) bool {
	fitActivity := NewFitActivity(fitActivityFilepath)
	activityTitle := fitActivity.ExtractActivityTitle()
	isTreadmill := fitActivity.IsTreadmill()
	s.logger.Infof("Activity Title: %s | Is Treadmill: %t\n", activityTitle, isTreadmill)

	token, err := s.Client.LoadAuthenticityToken(s.Client.EndpointForm)
	if err != nil {
		s.logger.Errorf("Error loading form requirements: %v\n", err)
		return false
	}
	// Waiting for 5 seconds before processing the next request...
	time.Sleep(5 * time.Second)

	s.logger.Info("Authenticity token for file upload found")

	uploadActivity, err := s.Client.UploadActivity(fitActivityFilepath, token)
	if err != nil {
		s.logger.Errorf("%v\n", err)
		return false
	}

	s.logger.Infof("Uploaded activity with progress ID: %d\n", uploadActivity.ID)
	return true
}
