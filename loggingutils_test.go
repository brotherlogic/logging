package main

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

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

}

func TestBadFullLoad(t *testing.T) {
	s := InitTestServer()
	s.saveLogs(context.Background(), "blahload", time.Now().Unix(), nil)
	s.test = true

	_, err := s.loadAllLogs(context.Background(), "blahload")
	if err == nil {
		t.Errorf("Did not fail")
	}

}
