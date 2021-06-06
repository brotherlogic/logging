package main

import (
	"context"
	"os"
	"testing"
	"time"

	pb "github.com/brotherlogic/logging/proto"
)

func InitTestServer() *Server {
	s := Init()
	os.RemoveAll(".test")
	s.path = ".test"
	s.dpath = "testdata/"
	return s
}

func TestBasicCall(t *testing.T) {
	s := InitTestServer()

	_, err := s.Log(context.Background(), &pb.LogRequest{Log: &pb.Log{Origin: "test", Timestamp: time.Now().Unix(), Ttl: 1}})

	if err != nil {
		t.Errorf("Error in logging: %v", err)
	}

	_, err = s.Log(context.Background(), &pb.LogRequest{Log: &pb.Log{Origin: "test", Timestamp: time.Now().Unix(), Ttl: 1}})

	if err != nil {
		t.Errorf("Error in logging: %v", err)
	}

	logs, err := s.GetLogs(context.Background(), &pb.GetLogsRequest{Origin: "test"})
	if err != nil {
		t.Errorf("Error in logging:%v", err)
	}

	if len(logs.GetLogs()) != 2 {
		t.Errorf("bad number of logs: (%v) %v", len(logs.GetLogs()), logs)
	}

	time.Sleep(time.Second * 5)

	_, err = s.Log(context.Background(), &pb.LogRequest{Log: &pb.Log{Origin: "test", Timestamp: time.Now().Unix(), Ttl: 1}})
	s.clean()

	logs, err = s.GetLogs(context.Background(), &pb.GetLogsRequest{Origin: "test"})
	if err != nil {
		t.Errorf("Error in logging:%v", err)
	}

	if len(logs.GetLogs()) != 1 {
		t.Errorf("bad number of logs: (%v) %v", len(logs.GetLogs()), logs)
	}
}

func TestDLogCall(t *testing.T) {
	s := InitTestServer()

	logs, err := s.GetLogs(context.Background(), &pb.GetLogsRequest{Origin: "testbin", IncludeDlogs: true})
	if err != nil {
		t.Errorf("Error getting logs: %v", err)
	}

	if len(logs.GetLogs()) != 1 {
		t.Fatalf("No logs read")
	}

	if logs.GetLogs()[0].Timestamp != 1136214245 {
		t.Errorf("Bad timestamp: %v", logs.GetLogs()[0])
	}
}

func TestDLogCallFail(t *testing.T) {
	s := InitTestServer()
	s.test = true

	logs, err := s.GetLogs(context.Background(), &pb.GetLogsRequest{Origin: "testbin", IncludeDlogs: true})
	if err == nil {
		t.Errorf("Should have failed: %v", logs)
	}
}
