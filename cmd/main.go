package main

import (
	"fmt"
	"path/filepath"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/mxdc/nrc2strava/converter"
	"github.com/mxdc/nrc2strava/fit"
	"github.com/mxdc/nrc2strava/migrator"
	"github.com/mxdc/nrc2strava/nrc"
	"github.com/mxdc/nrc2strava/parser"
	"github.com/mxdc/nrc2strava/strava"
	"github.com/mxdc/nrc2strava/utils"
	"github.com/sirupsen/logrus"
)

var (
	// migrate
	migrate               = kingpin.Command("migrate", "Migrate NRC activities to Strava.")
	migrateToken          = migrate.Flag("nrc.token", "NRC access token").Default("").String()
	migrateActivityDir    = migrate.Flag("fit.dir", "FIT activities directory").Default("").String()
	migrateStrava4Session = migrate.Flag("strava.token", "Strava session token").Default("").String()

	// download
	download              = kingpin.Command("download", "Download NRC activities.")
	downloadActivitiesDir = download.Flag("activities.dir", "Downloaded NRC activities directory").Default("./downloaded").String()
	downloadToken         = download.Flag("nrc.token", "NRC access token").Default("").String()

	// strava-download
	stravaDownload              = kingpin.Command("strava-download", "Download Strava activities.")
	stravaDownloadActivitiesDir = stravaDownload.Flag("activities.dir", "Downloaded Strava activities directory").Default("./strava-downloaded").String()
	stravaDownloadToken         = stravaDownload.Flag("strava.token", "Strava session token").Default("").String()

	// convert
	convert          = kingpin.Command("convert", "Convert NRC activities into FIT activities.")
	nrcActivitiesDir = convert.Flag("activities.dir", "Downloaded NRC activities directory").Default("").String()
	nrcActivityFile  = convert.Flag("activity.file", "Downloaded NRC Activity file").Default("").String()
	outputDir        = convert.Flag("fit.dir", "FIT Activities output directory").Default("./output").String()

	// upload
	upload                = kingpin.Command("upload", "Upload FIT activities to Strava.")
	uploadStrava4Session  = upload.Flag("strava.token", "Strava session token").Default("").String()
	uploadFitActivityFile = upload.Flag("fit.file", "FIT activity file").Default("").String()
	uploadFitActivityDir  = upload.Flag("fit.dir", "FIT activities directory").Default("").String()

	// merge
	merge           = kingpin.Command("merge", "Merge two FIT activities into one.")
	mergeFitFile1   = merge.Flag("fit.file1", "First FIT activity file").Required().String()
	mergeFitFile2   = merge.Flag("fit.file2", "Second FIT activity file").Required().String()
	mergeOutputFile = merge.Flag("output.file", "Output merged FIT file path").Default("").String()

	// logger
	logger = logrus.New()
)

func init() {
	kingpin.Parse()
	logger.SetFormatter(utils.LogFormat)
}

func main() {
	kingpin.Version("1.0.0")
	switch kingpin.Parse() {
	case migrate.FullCommand():
		handleMigrate(*migrateToken, *migrateStrava4Session, *migrateActivityDir)
	case download.FullCommand():
		handleDownload(*downloadActivitiesDir, *downloadToken)
	case convert.FullCommand():
		handleConvert(*nrcActivitiesDir, *nrcActivityFile, *outputDir)
	case upload.FullCommand():
		handleUpload(*uploadFitActivityDir, *uploadFitActivityFile, *uploadStrava4Session)
	case stravaDownload.FullCommand():
		handleStravaDownload(*stravaDownloadActivitiesDir, *stravaDownloadToken)
	case merge.FullCommand():
		handleMerge(*mergeFitFile1, *mergeFitFile2, *mergeOutputFile)
	default:
		kingpin.Usage()
	}
}

func handleMigrate(downloadToken, strava4Session, outputDir string) {
	nikeApi := nrc.NewNikeApi(downloadToken)
	stravaWeb := strava.NewStravaWeb(strava4Session)
	migrate := migrator.NewMigrator(nikeApi, stravaWeb, outputDir)
	migrate.MigrateActivities()
}

func handleDownload(downloadActivitiesDir, accessToken string) {
	if len(downloadActivitiesDir) == 0 {
		logger.Error("Please provide a directory to save the downloaded activities.")
		return
	}

	nikeApi := nrc.NewNikeApi(accessToken)
	nikeDownloader := nrc.NewNikeDownloader(nikeApi, downloadActivitiesDir)
	nikeDownloader.DownloadActivities()
}

func handleStravaDownload(stravaDownloadActivitiesDir, stravaDownloadToken string) {
	if len(stravaDownloadActivitiesDir) == 0 {
		logger.Error("Please provide a directory to save the downloaded activities.")
		return
	}

	stravaWeb := strava.NewStravaWeb(stravaDownloadToken)
	stravaDownloader := strava.NewStravaDownloader(stravaWeb, stravaDownloadActivitiesDir)
	stravaDownloader.DownloadActivities()
}

func handleUpload(fitActivityDir, fitActivityFile, strava4Session string) {
	if len(fitActivityDir) == 0 && len(fitActivityFile) == 0 {
		logger.Error("Please provide either a FIT activity file or a directory of FIT activities.")
		return
	}

	stravaWeb := strava.NewStravaWeb(strava4Session)
	stravaUploader := strava.NewStravaUploader(fitActivityFile, stravaWeb)

	if len(fitActivityFile) > 0 {
		logger.Infof("Processing file: %s\n", fitActivityFile)
		stravaUploader.UploadActivity(fitActivityFile)
	}

	if len(fitActivityDir) > 0 {
		mover := fit.NewActivityMover(filepath.Join(fitActivityDir, "uploaded"))
		batchUploader := strava.NewBatchUploader(stravaUploader, mover)

		if _, _, err := batchUploader.UploadDir(fitActivityDir); err != nil {
			logger.Errorf("Error reading directory: %v\n", err)
		}
	}
}

func handleConvert(activitiesDir, activityFile, outputDir string) {
	if len(activitiesDir) == 0 && len(activityFile) == 0 {
		logger.Error("Please provide either an activity file or a directory of activities.")
		return
	}

	activitiesParser := parser.NewActivitiesParser(activitiesDir, activityFile)
	activitiesConverter := converter.NewActivitiesConverter()
	activityWriter := fit.NewActivityWriter(outputDir)

	if len(activityFile) > 0 {
		nikeActivity := activitiesParser.LoadActivity()
		run := activitiesConverter.ConvertRun(nikeActivity)
		activityWriter.WriteFIT(run)
	}

	if len(activitiesDir) > 0 {
		nikeActivities := activitiesParser.LoadActivities()

		if len(nikeActivities) == 0 {
			logger.Error("No activities to convert")
			return
		}

		batchConverter := converter.NewBatchConverter(activitiesConverter, activityWriter)
		batchConverter.ConvertAll(nikeActivities)
	}
}

func handleMerge(fitFile1, fitFile2, outputFile string) {
	if len(fitFile1) == 0 || len(fitFile2) == 0 {
		logger.Error("Please provide both FIT activity files.")
		return
	}

	// Generate output filename if not provided
	if len(outputFile) == 0 {
		timestamp := time.Now().Format("20060102_150405")
		outputFile = fmt.Sprintf("./merged_%s.fit", timestamp)
	}

	// Create merger and execute
	merger := fit.NewActivityMerger(fitFile1, fitFile2, outputFile)
	err := merger.MergeActivities()
	if err != nil {
		logger.Errorf("Merge failed: %v\n", err)
		return
	}

	logger.Infof("✓ Successfully merged activities into: %s\n", outputFile)
}
