package main

import (
	"os"
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
	migrate                    = kingpin.Command("migrate", "Migrate NRC activities to Strava.")
	migrateToken               = migrate.Flag("nrc.token", "NRC access token").Default("").String()
	migrateActivityDir         = migrate.Flag("fit.dir", "FIT activities directory").Default("").String()
	migrateStrava4Session      = migrate.Flag("strava.token", "Strava session token").Default("").String()
	migrateXpSessionIdentifier = migrate.Flag("strava.id", "Strava session identifier").Default("").String()

	// download
	download              = kingpin.Command("download", "Download NRC activities.")
	downloadActivitiesDir = download.Flag("activities.dir", "Downloaded NRC activities directory").Default("./downloaded").String()
	downloadToken         = download.Flag("nrc.token", "NRC access token").Default("").String()

	// convert
	convert          = kingpin.Command("convert", "Convert NRC activities into FIT activities.")
	nrcActivitiesDir = convert.Flag("activities.dir", "Downloaded NRC activities directory").Default("").String()
	nrcActivityFile  = convert.Flag("activity.file", "Downloaded NRC Activity file").Default("").String()
	outputDir        = convert.Flag("fit.dir", "FIT Activities output directory").Default("./output").String()

	// upload
	upload                    = kingpin.Command("upload", "Upload FIT activities to Strava.")
	uploadStrava4Session      = upload.Flag("strava.token", "Strava session token").Default("").String()
	uploadXpSessionIdentifier = upload.Flag("strava.id", "Strava session identifier").Default("").String()
	uploadFitActivityFile     = upload.Flag("fit.file", "FIT activity file").Default("").String()
	uploadFitActivityDir      = upload.Flag("fit.dir", "FIT activities directory").Default("").String()

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
		handleMigrate(*migrateToken, *migrateStrava4Session, *migrateXpSessionIdentifier, *migrateActivityDir)
	case download.FullCommand():
		handleDownload(*downloadActivitiesDir, *downloadToken)
	case convert.FullCommand():
		handleConvert(*nrcActivitiesDir, *nrcActivityFile, *outputDir)
	case upload.FullCommand():
		handleUpload(*uploadFitActivityDir, *uploadFitActivityFile, *uploadStrava4Session, *uploadXpSessionIdentifier)
	default:
		kingpin.Usage()
	}
}

func handleMigrate(downloadToken, strava4Session, xpSessionIdentifier, outputDir string) {
	nikeApi := nrc.NewNikeApi(downloadToken)
	stravaWeb := strava.NewStravaWeb(strava4Session, xpSessionIdentifier)
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

func handleUpload(fitActivityDir, fitActivityFile, strava4Session, xpSessionIdentifier string) {
	if len(fitActivityDir) == 0 && len(fitActivityFile) == 0 {
		logger.Error("Please provide either a FIT activity file or a directory of FIT activities.")
		return
	}

	stravaWeb := strava.NewStravaWeb(strava4Session, xpSessionIdentifier)
	stravaUploader := strava.NewStravaUploader(fitActivityFile, stravaWeb)

	if len(fitActivityFile) > 0 {
		logger.Infof("Processing file: %s\n", fitActivityFile)
		stravaUploader.UploadActivity(fitActivityFile)
	}

	if len(fitActivityDir) > 0 {
		files, err := os.ReadDir(fitActivityDir)
		if err != nil {
			logger.Errorf("Error reading directory: %v\n", err)
			return
		}

		total := len(files)
		logger.Infof("Total file(s) to upload: %d\n", total)

		for index, file := range files {
			// Only process .fit files
			if filepath.Ext(file.Name()) == ".fit" {
				filePath := filepath.Join(fitActivityDir, file.Name())
				logger.Infof("Processing file: %s\n", filePath)

				success := stravaUploader.UploadActivity(filePath)
				if !success {
					return
				}

				// move the file to a different directory if upload is successful
				destinationDir := filepath.Join(fitActivityDir, "uploaded")
				fit.InitActivityMover(destinationDir).MoveFIT(filePath, file.Name())

				if index < total-1 {
					logger.Info("Waiting for 10 seconds before processing the next file...")
					time.Sleep(10 * time.Second)
				}
			}
		}
	}
}

func handleConvert(activitiesDir, activityFile, outputDir string) {
	if len(activitiesDir) == 0 && len(activityFile) == 0 {
		logger.Error("Please provide either an activity file or a directory of activities.")
		return
	}

	activitiesParser := parser.InitActivitiesParser(activitiesDir, activityFile)
	activitiesConverter := converter.InitActivitiesConverter()
	activityWriter := fit.InitActivityWriter(outputDir)

	if len(activityFile) > 0 {
		nikeActivity := activitiesParser.LoadActivity()
		run := activitiesConverter.ConvertRun(nikeActivity)
		activityWriter.WriteFIT(run)
	}

	if len(activitiesDir) > 0 {
		nikeActivities := activitiesParser.LoadActivities()
		logger.Infof("Converting %d activities\n", len(nikeActivities))

		for _, nikeActivity := range nikeActivities {
			run := activitiesConverter.ConvertRun(nikeActivity)
			activityWriter.WriteFIT(run)
		}
	}
}
