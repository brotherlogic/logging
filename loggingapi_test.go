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
