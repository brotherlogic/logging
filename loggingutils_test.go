package main

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestConvert(t *testing.T) {
	s := InitTestServer()
	val := "2022-06-20T19:13:28.861837314-07:00|clust2|recordcleaner-cli-1655777608-1633026440451240047-toru|S: /recordcleaner.RecordCleanerService/GetClean <-  bytes"
	conv := s.convert(strings.Split(val, "|"))

	if time.Unix(0, conv.GetTimestamp()).Format(time.RFC3339Nano) != "2022-06-20T19:13:28.861837314-07:00" {
		t.Errorf("Onversion issue: %v", time.Unix(conv.GetTimestamp(), 0).Format(time.RFC3339Nano))
	}
}

func TestBadSave(t *testing.T) {
	s := InitTestServer()
	s.test = true

	err := s.saveLogs(context.Background(), "blah", time.Now().Unix(), nil)
	if err == nil {
		t.Errorf("Did not fail")
	}
}

func TestBadLoad(t *testing.T) {
	s := InitTestServer()
	s.saveLogs(context.Background(), "blahload", time.Now().Unix(), nil)
	s.test = true

	_, err := s.loadLogs(context.Background(), "blahload", time.Now().Unix())
	if err == nil {
		t.Errorf("Did not fail")
	}

	err = s.cleanAllLogs()
	if err == nil {
		t.Errorf("Did not fail")
	}

}

func TestBadFullLoad(t *testing.T) {
	s := InitTestServer()
	s.saveLogs(context.Background(), "blahload", time.Now().Unix(), nil)
	s.test = true

	_, err := s.loadAllLogs(context.Background(), "blahload", "", false, "")
	if err == nil {
		t.Errorf("Did not fail")
	}

	err = s.cleanAllLogs()
	if err == nil {
		t.Errorf("Did not fail")
	}
}

func TestMin(t *testing.T) {
	if min(1, 10) != 1 || min(10, 1) != 1 {
		t.Errorf("Min is wrong")
	}
}
