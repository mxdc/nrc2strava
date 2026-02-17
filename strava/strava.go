package strava

import (
	"github.com/mxdc/nrc2strava/utils"
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
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &StravaUploader{
		FitActivityFile: fitActivityFile,
		Client:          stravaWeb,
		logger:          logger,
	}
}

func (s *StravaUploader) UploadActivity(fitActivityFilepath string) bool {
	fitActivity := NewFitActivity(fitActivityFilepath)
	activityTitle := fitActivity.ExtractActivityTitle()
	isTreadmill := fitActivity.IsTreadmill()
	s.logger.Debugf("Activity Title: %s | Is Treadmill: %t\n", activityTitle, isTreadmill)

	token, err := s.Client.LoadAuthenticityToken(s.Client.EndpointForm)
	if err != nil {
		s.logger.Errorf("Error loading form requirements: %v\n", err)
		return false
	}
	s.logger.Debug("Authenticity token for file upload found")

	uploadActivity, err := s.Client.UploadActivity(fitActivityFilepath, token)
	if err != nil {
		s.logger.Errorf("Upload error: %v\n", err)
		return false
	}

	s.logger.Debugf("Uploaded activity with progress ID: %d, and name: %s\n", uploadActivity.ID, activityTitle)
	return true
}
