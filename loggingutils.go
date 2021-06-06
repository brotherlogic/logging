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

func (s *Server) loadAllLogs(ctx context.Context, origin string, match string, includeDLogs bool) ([]*pb.Log, error) {
	logs := []*pb.Log{}

	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, origin) && !info.IsDir() {
			nlogs, err := s.loadLogFile(path)
			if err != nil {
				return err
			}
			for _, log := range nlogs {
				if match == "" || strings.Contains(log.GetLog(), match) {
					logs = append(logs, log)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Walk the dlogs if we've been asked to
	if includeDLogs {
		err = filepath.Walk(fmt.Sprintf("%v/%v", s.dpath, origin), func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, origin) && !info.IsDir() {
				dlogs, err := s.loadDLogFile(path)
				if err != nil {
					return err
				}
				for _, log := range dlogs {
					if match == "" || strings.Contains(log.GetLog(), match) {
						logs = append(logs, log)
					}
				}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].GetTimestamp() > logs[j].GetTimestamp()
	})

	// Only return 20 logs
	return logs[0:min(20, len(logs))], err
}

func (s *Server) cleanAllLogs() error {
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			nlogs, err := s.loadLogFile(path)
			if err != nil {
				return err
			}
			newlogs := []*pb.Log{}
			for _, log := range nlogs {
				if time.Now().Sub(time.Unix(log.GetTimestamp(), 0)).Seconds() < float64(log.GetTtl()) {
					newlogs = append(newlogs, log)
				}
			}
			data, err := s.marshal(newlogs)
			if err == nil {
				err = ioutil.WriteFile(path, data, 0644)
			}
			return err
		}
		return nil
	})

	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) loadLogs(ctx context.Context, origin string, timestamp int64) ([]*pb.Log, error) {
	fname, _ := s.getFileName(origin, timestamp)
	return s.loadLogFile(fname)
}

func (s *Server) loadLogFile(fname string) ([]*pb.Log, error) {
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

func (s *Server) loadDLogFile(fname string) ([]*pb.Log, error) {
	return s.loadDLog(fname)
}
