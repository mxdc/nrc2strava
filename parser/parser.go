package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mxdc/nrc2strava/types"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

// NikeActivitiesParser loads the activities files and parse them
type ActivitiesParser struct {
	ActivitiesDir string
	activityFile  string

	// logger
	logger *logrus.Logger
}

// InitActivitiesParser returns an initialized ActivitiesParser
func InitActivitiesParser(activitiesDir, activityFile string) *ActivitiesParser {
	var parser ActivitiesParser

	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	parser.ActivitiesDir = activitiesDir
	parser.activityFile = activityFile
	parser.logger = logger

	return &parser
}

// LoadActivities load JSON files into memory
func (p *ActivitiesParser) LoadActivities() []*types.Activity {
	p.logger.Debugf("Opening file at %s", p.ActivitiesDir)
	var activities []*types.Activity

	if len(p.ActivitiesDir) > 0 {
		activities = p.parseActivities()
	}

	return activities
}

func (p *ActivitiesParser) LoadActivity() *types.Activity {
	p.logger.Debugf("Opening file at %s", p.activityFile)

	if len(p.activityFile) > 0 {
		activity := p.parseActivity(p.activityFile)
		p.logger.Debugf("Activity ID: %s, Status: %s\n", activity.ID, activity.Status)
		return activity
	}

	return nil
}

func (p *ActivitiesParser) parseActivities() []*types.Activity {
	var activities []*types.Activity

	// Read all files in the folder
	files, err := os.ReadDir(p.ActivitiesDir)
	if err != nil {
		p.logger.Errorf("Error reading directory: %s", p.ActivitiesDir)
		return activities
	}

	// Count .json files
	jsonFiles := []os.DirEntry{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			jsonFiles = append(jsonFiles, file)
		}
	}

	if len(jsonFiles) == 0 {
		p.logger.Error("No JSON files to process")
		return activities
	}

	p.logger.Debugf("Processing %d activities...\n", len(jsonFiles))

	// Create progress bar
	bar := progressbar.NewOptions(len(jsonFiles),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetDescription("→ Parsing activities"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
	)

	// Loop through all files
	for _, file := range jsonFiles {
		filePath := filepath.Join(p.ActivitiesDir, file.Name())

		activity := p.parseActivity(filePath)
		if activity == nil {
			bar.Add(1)
			continue
		}

		activities = append(activities, activity)
		bar.Add(1)
	}

	bar.Finish()
	fmt.Printf("\n✓ Parsed %d running activities\n", len(activities))
	return activities
}

func (p *ActivitiesParser) parseActivity(filePath string) *types.Activity {
	p.logger.Debugf("Processing file: %s", filePath)

	// Open and read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		p.logger.Errorf("Error reading file: %v", err)
		return nil
	}

	// Unmarshal JSON into Go struct
	var activity types.Activity
	err = json.Unmarshal(data, &activity)
	if err != nil {
		p.logger.Errorf("Error parsing JSON: %v", err)
		return nil
	}

	return &activity
}
