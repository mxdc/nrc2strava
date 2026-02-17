package strava

import (
	"log"
	"os"
	"time"
)

// StravaUploader represents the Strava API client
type StravaUploader struct {
	FitActivityFile string
	Client          *StravaWeb

	// logger
	logger *log.Logger
}

// NewStravaUploader initializes a new StravaUploader instance
func NewStravaUploader(fitActivityFile string, stravaWeb *StravaWeb) *StravaUploader {
	return &StravaUploader{
		FitActivityFile: fitActivityFile,
		Client:          stravaWeb,
		logger:          log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (s *StravaUploader) UploadActivity(fitActivityFilepath string) {
	fitActivity := NewFitActivity(fitActivityFilepath)
	activityTitle := fitActivity.ExtractActivityTitle()
	isTreadmill := fitActivity.IsTreadmill()
	s.logger.Printf("Activity Title: %s | Is Treadmill: %t\n", activityTitle, isTreadmill)

	token, err := s.Client.LoadAuthenticityToken(s.Client.EndpointForm)
	if err != nil {
		s.logger.Printf("Error loading form requirements: %v\n", err)
		return
	}
	// Waiting for 5 seconds before processing the next request...
	time.Sleep(5 * time.Second)

	s.logger.Println("Authenticity token for file upload found")

	uploadActivity, err := s.Client.UploadActivity(fitActivityFilepath, token)
	if err != nil {
		s.logger.Printf("Error uploading activity: %v\n", err)
		return
	}
	s.logger.Printf("Uploaded activity with progress ID: %d\n", uploadActivity.ID)
}
