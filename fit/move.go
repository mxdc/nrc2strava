package fit

import (
	"log"
	"os"
	"path/filepath"
)

// ActivityMover move FIT files
type ActivityMover struct {
	destinationDir string

	// logger
	logger *log.Logger
}

// InitActivityMover returns an initialized InitActivityMover
func InitActivityMover(outputDir string) *ActivityMover {
	var mover ActivityMover

	mover.destinationDir = outputDir
	mover.logger = log.New(os.Stderr, "", log.LstdFlags)

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

	m.logger.Printf("Moved file to: %s\n", destination)
}
