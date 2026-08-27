package fit

import (
	"os"
	"path/filepath"

	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// ActivityMover move FIT files
type ActivityMover struct {
	destinationDir string

	// logger
	logger *logrus.Logger
}

// NewActivityMover returns an initialized NewActivityMover
func NewActivityMover(outputDir string) *ActivityMover {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &ActivityMover{
		destinationDir: outputDir,
		logger:         logger,
	}
}

// MoveFIT moves FIT files
func (m *ActivityMover) MoveFIT(source, filename string) {
	if err := os.MkdirAll(m.destinationDir, os.ModePerm); err != nil {
		m.logger.Fatalf("Error creating directory: %v", err)
	}

	destination := filepath.Join(m.destinationDir, filename)
	if err := os.Rename(source, destination); err != nil {
		m.logger.Fatalf("Error moving file: %v", err)
	}

	m.logger.Debugf("Moved file to: %s\n", destination)
}
