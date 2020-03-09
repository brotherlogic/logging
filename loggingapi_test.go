package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/brotherlogic/logging/proto"
)

func InitTestServer() *Server {
	s := Init()
	s.path = ".test"
	return s
}

func TestBasicCall(t *testing.T) {
	s := InitTestServer()

	_, err := s.Log(context.Background(), &pb.LogRequest{Log: &pb.Log{Origin: "test", Timestamp: time.Now().Unix()}})

	if err != nil {
		t.Errorf("Error in logging: %v", err)
	}

	_, err = s.Log(context.Background(), &pb.LogRequest{Log: &pb.Log{Origin: "test", Timestamp: time.Now().Unix()}})

	if err != nil {
		t.Errorf("Error in logging: %v", err)
	}

}
