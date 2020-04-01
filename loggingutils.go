package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func (s *Server) loadAllLogs(ctx context.Context, origin string) ([]*pb.Log, error) {
	logs := []*pb.Log{}

	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, origin) && !info.IsDir() {
			nlogs, err := s.loadLogFile(ctx, path)
			if err != nil {
				return err
			}
			logs = append(logs, nlogs...)
		}
		return nil
	})

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].GetTimestamp() < logs[j].GetTimestamp()
	})

	// Only return 20 logs
	return logs[0:min(20, len(logs))], err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) loadLogs(ctx context.Context, origin string, timestamp int64) ([]*pb.Log, error) {
	fname, _ := s.getFileName(origin, timestamp)
	return s.loadLogFile(ctx, fname)
}

func (s *Server) loadLogFile(ctx context.Context, fname string) ([]*pb.Log, error) {
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
