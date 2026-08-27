package converter

import (
	"github.com/mxdc/nrc2strava/fit"
	"github.com/mxdc/nrc2strava/types"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

// BatchConverter converts a batch of activities to FIT, writing each to disk.
type BatchConverter struct {
	converter *ActivitiesConverter
	writer    *fit.ActivityWriter

	// logger
	logger *logrus.Logger
}

// NewBatchConverter returns an initialized BatchConverter
func NewBatchConverter(converter *ActivitiesConverter, writer *fit.ActivityWriter) *BatchConverter {
	logger := logrus.New()
	logger.SetFormatter(utils.LogFormat)

	return &BatchConverter{
		converter: converter,
		writer:    writer,
		logger:    logger,
	}
}

// ConvertAll converts and writes every activity, returning how many were processed.
func (b *BatchConverter) ConvertAll(activities []*types.Activity) int {
	b.logger.Infof("Converting %d activities...\n", len(activities))

	for _, activity := range activities {
		run := b.converter.ConvertRun(activity)
		b.writer.WriteFIT(run)
	}

	b.logger.Infof("✓ Finished converting %d activities\n", len(activities))
	return len(activities)
}
