package strava

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/muktihari/fit/decoder"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// StravaDownloader represents the Strava API client
type StravaDownloader struct {
	downloadActivitiesDir string
	stravaWeb             *StravaWeb
	logger                *logrus.Logger
}

// NewStravaDownloader initializes a new NewStravaDownloader instance
func NewStravaDownloader(stravaWeb *StravaWeb, downloadActivitiesDir string) *StravaDownloader {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &StravaDownloader{
		downloadActivitiesDir: downloadActivitiesDir,
		stravaWeb:             stravaWeb,
		logger:                logger,
	}
}

func (s *StravaDownloader) DownloadActivities() {
	s.logger.Info("Downloading activities from Strava...")

	// Create the directory if it doesn't exist
	if _, err := os.Stat(s.downloadActivitiesDir); os.IsNotExist(err) {
		if err := os.Mkdir(s.downloadActivitiesDir, os.ModePerm); err != nil {
			s.logger.Errorf("Error creating activities folder: %v\n", err)
			return
		}
	}

	activities, err := s.stravaWeb.GetActivityList()
	if err != nil {
		s.logger.Errorf("Error fetching activity list: %v\n", err)
		return
	}

	total := len(activities)
	downloadedCount := 0

	for _, activity := range activities {
		// Build the download URL
		downloadURL := activity.ActivityURL + "/export_original"
		s.logger.Debugf("Downloading from: %s\n", downloadURL)

		// Download the file
		req, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			s.logger.Errorf("Error creating request for activity %d: %v\n", activity.ID, err)
			continue
		}

		// Add cookies
		cookies := []http.Cookie{
			{Name: "_strava4_session", Value: s.stravaWeb.Strava4Session},
		}
		for _, cookie := range cookies {
			req.AddCookie(&cookie)
		}

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			s.logger.Errorf("Error downloading activity %d: %v\n", activity.ID, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.logger.Errorf("Error downloading activity %d: status %s\n", activity.ID, resp.Status)
			continue
		}

		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			s.logger.Errorf("Error reading response body for activity %d: %v\n", activity.ID, err)
			continue
		}

		// Create a new FIT decoder
		d := decoder.New(bytes.NewReader(bodyBytes))

		// Decode the FIT file
		_, err = d.Decode()
		if err != nil {
			s.logger.Errorf("Failed to decode FIT response: %v", err)
			continue
		}

		// Format the filename: 2025-01-30_<ID>_<activity name>.fit
		activityDate := time.Unix(activity.StartDateLocalRaw, 0).Format("2006-01-02")
		sanitizedName := sanitizeFilename(activity.Name)
		finalFilename := fmt.Sprintf("%s_%d_%s.fit", activityDate, activity.ID, sanitizedName)

		// Save the file
		filePath := filepath.Join(s.downloadActivitiesDir, finalFilename)
		err = os.WriteFile(filePath, bodyBytes, 0644)
		if err != nil {
			s.logger.Errorf("Error saving activity %d: %v\n", activity.ID, err)
			continue
		}

		downloadedCount++
		s.logger.Infof("✓ Downloaded %d/%d activities (%s)\n", downloadedCount, total, finalFilename)
		time.Sleep(10 * time.Millisecond)
	}

	s.logger.Infof("✓ Finished downloading %d activities\n", downloadedCount)
}

// sanitizeFilename replaces spaces and invisible characters with underscores
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, "_")
	// Replace any non-alphanumeric characters (except hyphen, underscore) with underscores
	name = regexp.MustCompile(`[^\w\-]`).ReplaceAllString(name, "_")
	// Remove consecutive underscores
	name = regexp.MustCompile(`_+`).ReplaceAllString(name, "_")
	// Remove leading/trailing underscores
	name = regexp.MustCompile(`^_+|_+$`).ReplaceAllString(name, "")
	return name
}
