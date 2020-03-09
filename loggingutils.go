package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	pb "github.com/brotherlogic/logging/proto"
)

func (s *Server) getFileName(origin string, timestamp int64) (string, string) {
	t := time.Unix(timestamp, 0)
	return fmt.Sprintf("%v/%v/%v-%v-%v-%v.logs",
			s.path,
			origin,
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour()),
		fmt.Sprintf("%v/%v",
			s.path,
			origin)

}

func (s *Server) saveLogs(ctx context.Context, origin string, timestamp int64, logs []*pb.Log) error {
	fname, dir := s.getFileName(origin, timestamp)
	os.MkdirAll(dir, 0777)

	data, err := s.marshal(logs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, data, 0644)
}

func (s *Server) loadLogs(ctx context.Context, origin string, timestamp int64) ([]*pb.Log, error) {
	fname, _ := s.getFileName(origin, timestamp)

	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return []*pb.Log{}, nil
	}

	data, err := s.load(fname)
	if err != nil {
		return nil, err
	}

	list := &pb.LogList{}
	proto.Unmarshal(data, list)
	return list.GetLogs(), nil
}
