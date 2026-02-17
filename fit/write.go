package fit

import (
	"os"
	"path/filepath"

	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/mxdc/nrc2strava/types"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// ActivityWriter write FIT files
type ActivityWriter struct {
	OutputDir string

	// logger
	logger *logrus.Logger
}

// InitActivityWriter returns an initialized InitActivityWriter
func InitActivityWriter(outputDir string) *ActivityWriter {
	var writer ActivityWriter

	writer.OutputDir = outputDir
	writer.logger = logrus.New()
	writer.logger.SetFormatter(utils.LogFormat)

	return &writer
}

// LoadActivities load JSON files into memory
func (w *ActivityWriter) WriteFIT(run types.Run) string {
	// Ensure the output directory exists
	if err := os.MkdirAll(w.OutputDir, os.ModePerm); err != nil {
		panic(err)
	}

	// Convert back to FIT protocol messages
	fit := run.Activity.ToFIT(nil)

	outputFilename := w.generateFilename(run)

	w.logger.Infof("Writing file at %s", outputFilename)

	f, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		w.logger.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	enc := encoder.New(f)
	if err := enc.Encode(&fit); err != nil {
		w.logger.Fatalf("Error encoding FIT file: %v", err)
	}

	return outputFilename
}

func (w *ActivityWriter) generateFilename(run types.Run) string {
	parsedTime := run.Activity.Activity.Timestamp

	date := parsedTime.Format("2006-01-02")

	suffix := "outside"
	if run.Activity.Sessions[0].SubSport == typedef.SubSportTreadmill {
		suffix = "indoors"
	}

	outputFilename := date + "_" + suffix + "_" + run.Id + ".fit"

	// Combine the output directory and the filename
	fullPath := filepath.Join(w.OutputDir, outputFilename)

	return fullPath
}
