package strava

import (
	"log"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
)

type FitActivity struct {
	Fit *proto.FIT

	// logger
	logger *log.Logger
}

// NewFitActivity initializes a new FitActivity instance
func NewFitActivity(fitActivityFilepath string) *FitActivity {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// Open the .fit file
	file, err := os.Open(fitActivityFilepath)
	if err != nil {
		logger.Fatalf("Failed to open FIT file: %v", err)
	}
	defer file.Close()

	// Create a new FIT decoder
	d := decoder.New(file)

	// Decode the FIT file
	fit, err := d.Decode()
	if err != nil {
		logger.Fatalf("Failed to decode FIT file: %v", err)
	}

	return &FitActivity{Fit: fit, logger: logger}
}

func (f *FitActivity) IsTreadmill() bool {
	// Iterate through the decoded messages
	for _, mesg := range f.Fit.Messages {
		// Check if the message is of type Session
		if mesg.Num == typedef.MesgNumSession {
			// Parse the message into a Session struct
			session := mesgdef.NewSession(&mesg)

			if session.SubSport == typedef.SubSportTreadmill {
				return true
			}
		}
	}

	return false
}

func (f *FitActivity) ExtractActivityTitle() string {
	// Iterate through the decoded messages
	for _, mesg := range f.Fit.Messages {
		// Check if the message is of type FileId
		if mesg.Num == typedef.MesgNumFileId {
			// Parse the message into a FileId struct
			fileId := mesgdef.NewFileId(&mesg)

			if len(fileId.UnknownFields) > 0 {
				customField := fileId.UnknownFields[0]
				return customField.Value.String()
			}
		}
	}

	return "Imported from NRC"
}
