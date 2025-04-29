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

// InitActivityMover returns an initialized InitActivityMover
func InitActivityMover(outputDir string) *ActivityMover {
	var mover ActivityMover

	mover.destinationDir = outputDir
	mover.logger = logrus.New()
	mover.logger.SetFormatter(utils.LogFormat)

	return &mover
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

	m.logger.Infof("Moved file to: %s\n", destination)
}
