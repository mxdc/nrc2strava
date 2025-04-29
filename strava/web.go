package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// StravaWeb represents the Strava Web client
type StravaWeb struct {
	// Cookie Data | Domain: www.strava.com
	XpSessionIdentifier string
	Strava4Session      string

	// Endpoint
	EndpointForm   string
	EndpointUpload string

	// logger
	logger *logrus.Logger
}

// NewStravaWeb initializes a new StravaWeb instance
func NewStravaWeb(strava4Session, xpSessionIdentifier string) *StravaWeb {
	logger := logrus.New()

	return &StravaWeb{
		// Cookie Data | Domain: www.strava.com
		Strava4Session:      strava4Session,
		XpSessionIdentifier: xpSessionIdentifier,

		// Endpoint
		EndpointForm:   "https://www.strava.com/upload/select",
		EndpointUpload: "https://www.strava.com/upload/files",

		// logger
		logger: logger,
	}
}

// LoadAuthenticityToken performs a GET request and extracts the authenticity token from the HTML response
func (web *StravaWeb) LoadAuthenticityToken(endpoint string) (string, error) {
	web.logger.Infof("Loading authenticity token from: %s\n", endpoint)

	// Create the HTTP request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Add cookies
	cookies := []http.Cookie{
		{Name: "xp_session_identifier", Value: web.XpSessionIdentifier},
		{Name: "_strava4_session", Value: web.Strava4Session},
	}

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Print the response status
	web.logger.Infof("Response status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error loading authenticity token: %s", resp.Status)
	}

	// Parse the HTML response to extract the authenticity token
	token, err := extractAuthenticityToken(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error extracting authenticity token: %w", err)
	}

	return token, nil
}

// extractAuthenticityToken parses the HTML and extracts the authenticity token
func extractAuthenticityToken(body io.Reader) (string, error) {
	// Parse the HTML document
	doc, err := html.Parse(body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	// Traverse the HTML nodes to find the <input name="authenticity_token">
	var AuthenticityToken string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			// Check if the input element has the name="authenticity_token" attribute
			var isAuthToken bool
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "authenticity_token" {
					isAuthToken = true
				}
				if isAuthToken && attr.Key == "value" {
					AuthenticityToken = attr.Val
					return
				}
			}
		}

		// Recursively traverse child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	// Check if the token was found
	if AuthenticityToken == "" {
		return "", fmt.Errorf("authenticity_token not found in HTML")
	}

	// Return the extracted token
	return AuthenticityToken, nil
}

type UploadedActivity struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Progress  int    `json:"progress"`
	Workflow  string `json:"workflow"`
	StartDate string `json:"start_date"`
	Error     string `json:"error"`
}

type Activity struct {
	ID                int64   `json:"id"`
	Name              string  `json:"name"`
	StartDateLocalRaw int64   `json:"start_date_local_raw"`
	DistanceRaw       float64 `json:"distance_raw"`
	ShortUnit         string  `json:"short_unit"`
	MovingTimeRaw     int64   `json:"moving_time_raw"`
	ElapsedTimeRaw    int64   `json:"elapsed_time_raw"`
}

type ActivitiesResponse struct {
	Models  []Activity `json:"models"`
	Page    int        `json:"page"`
	PerPage int        `json:"perPage"`
	Total   int        `json:"total"`
}

func (web *StravaWeb) UploadActivity(filePath, token string) (*UploadedActivity, error) {
	web.logger.Infof("Uploading activity file: %s\n", filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file to the form
	part, err := writer.CreateFormFile("files[]", filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file content: %w", err)
	}

	// Add additional form fields
	_ = writer.WriteField("_method", "post")
	_ = writer.WriteField("authenticity_token", token)

	// Close the writer to finalize the form
	writer.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", web.EndpointUpload, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("content-type", writer.FormDataContentType())
	req.Header.Set("origin", "https://www.strava.com")
	req.Header.Set("referer", web.EndpointForm)
	req.Header.Set("x-csrf-token", token)

	// Add cookies using req.AddCookie
	cookies := []http.Cookie{
		{Name: "xp_session_identifier", Value: web.XpSessionIdentifier},
		{Name: "_strava4_session", Value: web.Strava4Session},
	}

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Print the response status
	web.logger.Infof("Response status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error uploading activity: %s", resp.Status)
	}

	// Read and parse the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the JSON response into a struct
	var response []UploadedActivity
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON response: %w", err)
	}

	if len(response) > 0 {
		return &response[0], nil
	}

	return nil, fmt.Errorf("no activity uploaded")
}
