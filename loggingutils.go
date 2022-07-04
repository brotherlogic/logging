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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"

	pb "github.com/brotherlogic/logging/proto"
)

var (
	//DirSize - the print queue
	filtered = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "logging_filtered",
		Help: "The size of the logs",
	})
	original = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "logging_original",
		Help: "The size of the logs",
	})
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
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	data, err := s.marshal(logs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, data, 0644)
}

func (s *Server) loadAllLogs(ctx context.Context, origin string, match string, includeDLogs bool, context string) ([]*pb.Log, error) {
	logs := []*pb.Log{}

	if !includeDLogs {
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
	}

	// Walk the dlogs if we've been asked to
	if includeDLogs {
		err := filepath.Walk(fmt.Sprintf("%v/%v", s.dpath, origin), func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if (origin == "" || strings.Contains(path, origin)) && !info.IsDir() {
					dlogs, err := s.loadDLogFile(path, origin, context)
					if err != nil {
						return err
					}
					for _, log := range dlogs {
						if match == "" || strings.Contains(log.GetLog(), match) {
							logs = append(logs, log)
						}
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

	// Filter by context if need to
	var nlogs []*pb.Log
	for _, log := range logs {
		if context == "" || log.GetContext() == context {
			nlogs = append(nlogs, log)
		}
	}

	original.Set(float64(len(logs)))
	filtered.Set(float64(len(nlogs)))

	// Only return 20 logs
	return nlogs[0:min(20, len(nlogs))], nil
}

func (s *Server) cleanAllLogs() error {
	var toDelete []string
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			s.DLog(context.Background(), fmt.Sprintf("Cleaning %v", path))
			nlogs, err := s.loadLogFile(path)
			if err != nil {
				return err
			}
			newlogs := []*pb.Log{}
			for _, log := range nlogs {
				if time.Since(time.Unix(0, log.GetTimestamp())).Seconds() < float64(log.GetTtl()) && time.Since(time.Unix(0, log.GetTimestamp())).Seconds() > 0 {
					s.DLog(context.Background(), fmt.Sprintf("%v -> %v from %v", time.Since(time.Unix(0, log.GetTimestamp())).Seconds(), time.Unix(0, log.GetTimestamp()), log))
					newlogs = append(newlogs, log)
				}
			}
			if len(newlogs) > 0 {
				data, err := s.marshal(newlogs)
				if err == nil {
					err = ioutil.WriteFile(path, data, 0644)
					return err
				}
			} else {
				// We can delete this file
				toDelete = append(toDelete, path)

			}
		}
		return nil
	})

	for _, td := range toDelete {
		os.Remove(td)
	}

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

func (s *Server) loadDLogFile(fname, origin, context string) ([]*pb.Log, error) {
	return s.loadDLog(fname, origin, context)
}
