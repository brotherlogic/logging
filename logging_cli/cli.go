package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/logging/proto"
)

func main() {
	ctx, cancel := utils.ManualContext("logging-cli", time.Second*10)
	defer cancel()
	servers, err := utils.LFFind(ctx, "logging")

	logFlags := flag.NewFlagSet("AddRecords", flag.ExitOnError)
	var include = logFlags.Bool("dlog", false, "Include dlogs")
	var matcher = logFlags.String("match", "", "Search string")
	var context = logFlags.String("context", "", "Context to search for")

	if err != nil {
		log.Fatalf("Error finding logging servers: %v", err)
	}

	logs := []*pb.Log{}
	for _, server := range servers {
		ctx, cancel := utils.ManualContext("logging-cli", time.Minute)
		defer cancel()
		conn, err := utils.LFDial(server)
		if err != nil {
			fmt.Printf("Dial error: %v", err)
			continue
		}
		defer conn.Close()
		client := pb.NewLoggingServiceClient(conn)

		var res *pb.GetLogsResponse

		if err := logFlags.Parse(os.Args[2:]); err == nil {
			res, err = client.GetLogs(ctx, &pb.GetLogsRequest{Origin: os.Args[1], Match: *matcher, IncludeDlogs: *include, Context: *context})
			if err == nil {
				logs = append(logs, res.GetLogs()...)
			}
		}

	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].GetTimestamp() > logs[j].GetTimestamp()
	})

	for _, l := range logs {
		fmt.Printf("%v - %v\n", time.Unix(l.GetTimestamp(), 0), l)
	}
}
