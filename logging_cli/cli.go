package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/logging/proto"
)

func main() {
	ctx, cancel := utils.ManualContext("logging-cli", "logging", time.Second*10, false)
	defer cancel()
	servers, err := utils.LFFind(ctx, "logging")

	if err != nil {
		log.Fatalf("Error finding logging servers: %v", err)
	}

	logs := []*pb.Log{}
	for _, server := range servers {
		ctx, cancel := utils.ManualContext("logging-cli", "logging", time.Minute, false)
		defer cancel()
		conn, err := utils.LFDial(server)
		if err != nil {
			fmt.Printf("Dial error: %v", err)
			continue
		}
		defer conn.Close()
		client := pb.NewLoggingServiceClient(conn)

		var res *pb.GetLogsResponse
		matcher := ""
		if len(os.Args) > 2 {
			matcher = os.Args[2]
		}

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
