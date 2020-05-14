package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/logging/proto"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	conn, err := grpc.Dial("discovery:///logging", grpc.WithInsecure(), grpc.WithBalancerName("my_pick_first"))
	if err != nil {
		log.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()

	client := pb.NewLoggingServiceClient(conn)
	ctx, cancel := utils.BuildContext("logging-cli", "logging")
	defer cancel()

	err = nil
	var res *pb.GetLogsResponse
	logs := []*pb.Log{}
	matcher := ""
	if len(os.Args) > 2 {
		matcher = os.Args[2]
	}
	for err == nil || status.Convert(err).Code() == codes.FailedPrecondition {
		res, err = client.GetLogs(ctx, &pb.GetLogsRequest{Origin: os.Args[1], Match: matcher})
		if err == nil {
			logs = append(logs, res.GetLogs()...)
		}
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].GetTimestamp() > logs[j].GetTimestamp()
	})

	for _, l := range logs {
		fmt.Printf("%v - %v\n", time.Unix(l.GetTimestamp(), 0), l)
	}
}
