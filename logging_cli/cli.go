package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"

	pb "github.com/brotherlogic/logging/proto"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	conn, err := grpc.Dial("discovery:///recordadder", grpc.WithInsecure(), grpc.WithBalancerName("my_pick_first"))
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
	for err == nil {
		res, err = client.GetLogs(ctx, &pb.GetLogsRequest{Origin: os.Args[1]})
		if err == nil {
			logs = append(logs, res.GetLogs()...)
		}
	}

	for _, l := range logs {
		fmt.Printf("%v\n", l)
	}
}
