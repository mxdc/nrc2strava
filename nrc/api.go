package nrc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// NikeApi represents the Nike API client
type NikeApi struct {
	ActivityListURL        string
	ActivityListPagination string
	ActivityDetailsURL     string
	AccessToken            string
	logger                 *logrus.Logger
}

// NewNikeApi initializes a new NikeApi instance
func NewNikeApi(accessToken string) *NikeApi {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &NikeApi{
		ActivityListURL:    "https://api.nike.com/plus/v3/activities/before_id/v3",
		ActivityDetailsURL: "https://api.nike.com/sport/v3/me/activity/%s?metrics=ALL",
		AccessToken:        accessToken,
		logger:             logger,
	}
}

type ActivitiesListResponse struct {
	Activities []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Tags struct {
			RunType string `json:"com.nike.running.runtype"`
		} `json:"tags"`
	} `json:"activities"`
	Paging struct {
		BeforeID string `json:"before_id"`
	} `json:"paging"`
}

func (n *NikeApi) buildActivityListURL(beforeID string, params url.Values) (*url.URL, error) {
	baseURL, err := url.Parse(fmt.Sprintf("%s/%s", n.ActivityListURL, "*"))
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %w", err)
	}

	// Add pagination parameter if applicable
	if beforeID != "" {
		baseURL, err = url.Parse(fmt.Sprintf("%s/%s", n.ActivityListURL, beforeID))
		if err != nil {
			return nil, fmt.Errorf("error parsing paginated URL: %w", err)
		}
	}

	// Attach query parameters
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func (n *NikeApi) fetchActivityList(url string) (*ActivitiesListResponse, error) {
	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.AccessToken))

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response ActivitiesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response, nil
}

func (n *NikeApi) GetActivityList() ([]string, error) {
	n.logger.Info("Collecting activities from Nike API...")

	var activityIDs []string
	beforeID := ""

	// Query parameters
	params := url.Values{}
	params.Set("limit", "30")
	params.Set("types", "run,jogging")
	params.Set("include_deleted", "false")

	for {
		// Build the URL with pagination
		baseURL, err := n.buildActivityListURL(beforeID, params)
		if err != nil {
			return nil, fmt.Errorf("error building URL: %w", err)
		}

		n.logger.Debugf("Opening page: %s\n", baseURL.String())

		// Make the HTTP request
		response, err := n.fetchActivityList(baseURL.String())
		if err != nil {
			return nil, fmt.Errorf("error fetching activity list: %w", err)
		}

		// Process activities
		for _, activity := range response.Activities {
			if activity.Type == "run" && activity.Tags.RunType != "manual" {
				activityIDs = append(activityIDs, activity.ID)
			}
		}

		// Log progress
		n.logger.Infof("✓ Collected %d activities\n", len(activityIDs))

		// Check for pagination
		if response.Paging.BeforeID == "" {
			break
		}

		beforeID = response.Paging.BeforeID
	}

	n.logger.Infof("✓ Finished collecting %d running activities\n", len(activityIDs))
	return activityIDs, nil
}

func (n *NikeApi) GetActivityDetailsWithRetry(activityID string, maxRetries int) ([]byte, error) {
	var body []byte
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		n.logger.Debugf("Attempt %d to fetch activity details for %s\n", attempt, activityID)

		body, err = n.GetActivityDetails(activityID)
		if err == nil {
			return body, nil
		}

		n.logger.Warnf("Error fetching activity details (attempt %d): %v\n", attempt, err)
		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("failed to fetch activity details for %s after %d attempts: %w", activityID, maxRetries, err)
}

func (n *NikeApi) GetActivityDetails(activityID string) ([]byte, error) {
	// Construct the request
	url := fmt.Sprintf(n.ActivityDetailsURL, activityID)
	n.logger.Debugf("New GET Request on: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.AccessToken))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Print the response status
	n.logger.Debugf("Response status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting activity details: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	n.logger.Debugf("Successfully fetched details for activity ID: %s\n", activityID)
	return body, nil
}
