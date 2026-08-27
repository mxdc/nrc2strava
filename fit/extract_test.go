package fit

import (
	"testing"

	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
)

func TestExtractFunctions(t *testing.T) {
	fitFile := &proto.FIT{
		Messages: []proto.Message{
			{Num: typedef.MesgNumFileId},
			{Num: typedef.MesgNumSession},
			{Num: typedef.MesgNumRecord},
			{Num: typedef.MesgNumRecord},
			{Num: typedef.MesgNumEvent},
			{Num: typedef.MesgNumEvent},
			{Num: typedef.MesgNumEvent},
			{Num: typedef.MesgNumLap},
			{Num: typedef.MesgNumActivity},
			{Num: typedef.MesgNumDeveloperDataId},
			{Num: typedef.MesgNumFieldDescription},
			{Num: typedef.MesgNumFieldDescription},
		},
	}

	if s := extractSession(fitFile); s == nil {
		t.Error("expected a session, got nil")
	}
	if f := extractFileId(fitFile); f == nil {
		t.Error("expected a FileId, got nil")
	}
	if r := extractRecords(fitFile); len(r) != 2 {
		t.Errorf("expected 2 records, got %d", len(r))
	}
	if e := extractEvents(fitFile); len(e) != 3 {
		t.Errorf("expected 3 events, got %d", len(e))
	}
	if l := extractLaps(fitFile); len(l) != 1 {
		t.Errorf("expected 1 lap, got %d", len(l))
	}
	if a := extractActivity(fitFile); a == nil {
		t.Error("expected an activity, got nil")
	}
	if ids := extractDeveloperDataIds(fitFile); len(ids) != 1 {
		t.Errorf("expected 1 developer data id, got %d", len(ids))
	}
	if fds := extractFieldDescriptions(fitFile); len(fds) != 2 {
		t.Errorf("expected 2 field descriptions, got %d", len(fds))
	}
}

func TestExtractSession_ReturnsNilWhenAbsent(t *testing.T) {
	fitFile := &proto.FIT{Messages: []proto.Message{{Num: typedef.MesgNumRecord}}}
	if s := extractSession(fitFile); s != nil {
		t.Error("expected nil session when none present")
	}
}

func TestExtractRecords_ReturnsNilWhenAbsent(t *testing.T) {
	fitFile := &proto.FIT{Messages: []proto.Message{{Num: typedef.MesgNumSession}}}
	if r := extractRecords(fitFile); len(r) != 0 {
		t.Errorf("expected 0 records, got %d", len(r))
	}
}
