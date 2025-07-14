package main

import (
	"golang.org/x/net/context"

	pb "github.com/brotherlogic/logging/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	request = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "logging_requests",
		Help: "The size of the logs",
	}, []string{"origin"})

	logSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "logging_logsize",
	}, []string{"orgin"})
)

// Log logs a message
func (s *Server) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	request.With(prometheus.Labels{"origin": req.GetLog().GetOrigin()}).Inc()
	logs, err := s.loadLogs(ctx, req.GetLog().GetOrigin(), req.GetLog().GetTimestamp())
	logs = append(logs, req.GetLog())
	logSize.With(prometheus.Labels{"origin": req.GetLog().GetOrigin()}).Set(float64(len(logs)))
	err = s.saveLogs(ctx, req.GetLog().GetOrigin(), req.GetLog().GetTimestamp(), logs)
	return &pb.LogResponse{}, err
}

// GetLogs gets the logs
func (s *Server) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	logs, err := s.loadAllLogs(ctx, req.GetOrigin(), req.GetMatch(), req.GetIncludeDlogs(), req.GetContext())
	return &pb.GetLogsResponse{Logs: logs}, err
}
