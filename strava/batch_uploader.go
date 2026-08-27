package strava

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mxdc/nrc2strava/fit"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// BatchUploader uploads every FIT file in a directory, moving each on success.
type BatchUploader struct {
	uploader *StravaUploader
	mover    *fit.ActivityMover

	// logger
	logger *logrus.Logger
}

// NewBatchUploader returns an initialized BatchUploader
func NewBatchUploader(uploader *StravaUploader, mover *fit.ActivityMover) *BatchUploader {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &BatchUploader{
		uploader: uploader,
		mover:    mover,
		logger:   logger,
	}
}

// UploadDir uploads every .fit file in dir, moving each on success.
// Stops at the first failure.
func (b *BatchUploader) UploadDir(dir string) (uploaded, total int, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading directory: %w", err)
	}

	// Count .fit files
	fitFiles := []os.DirEntry{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".fit" {
			fitFiles = append(fitFiles, file)
		}
	}

	total = len(fitFiles)
	if total == 0 {
		b.logger.Error("No .fit files to upload")
		return 0, 0, nil
	}

	b.logger.Infof("Uploading %d activities...\n", total)

	for _, file := range fitFiles {
		filePath := filepath.Join(dir, file.Name())
		b.logger.Debugf("Uploading file: %s\n", filePath)

		success := b.uploader.UploadActivity(filePath)
		if !success {
			return uploaded, total, nil
		}

		b.mover.MoveFIT(filePath, file.Name())

		uploaded++
		b.logger.Infof("✓ Uploaded %d/%d activities\n", uploaded, total)
		time.Sleep(100 * time.Millisecond)
	}

	b.logger.Infof("✓ Finished uploading %d activities\n", uploaded)
	return uploaded, total, nil
}
