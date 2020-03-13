package main

import "golang.org/x/net/context"

import pb "github.com/brotherlogic/logging/proto"

//Log logs a message
func (s *Server) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	logs, err := s.loadLogs(ctx, req.GetLog().GetOrigin(), req.GetLog().GetTimestamp())
	logs = append(logs, req.GetLog())
	err = s.saveLogs(ctx, req.GetLog().GetOrigin(), req.GetLog().GetTimestamp(), logs)
	return &pb.LogResponse{}, err
}

//GetLogs gets the logs
func (s *Server) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	logs, err := s.loadAllLogs(ctx, req.GetOrigin())
	return &pb.GetLogsResponse{Logs: logs}, err
}
