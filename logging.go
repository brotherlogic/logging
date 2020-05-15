package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/logging/proto"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"

	pbg "github.com/brotherlogic/goserver/proto"
)

func init() {
	resolver.Register(&utils.DiscoveryServerResolverBuilder{})
}

var (
	//Backlog - the print queue
	DirSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "logging_dirsize",
		Help: "The size of the logs",
	})
)

//Server main server type
type Server struct {
	*goserver.GoServer
	path    string
	test    bool
	dirSize int64
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
		path:     "/media/scratch/logs",
		test:     false,
		dirSize:  0,
	}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterLoggingServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

//Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "dir_size", Value: s.dirSize},
	}
}

func (s *Server) marshal(logs []*pb.Log) ([]byte, error) {
	data, err := proto.Marshal(&pb.LogList{Logs: logs})
	if err != nil {
		return []byte{}, err
	}
	if s.test {
		return []byte{}, fmt.Errorf("Testing failure")
	}
	return data, err
}

func (s *Server) load(fname string) ([]byte, error) {
	if s.test {
		return []byte{}, fmt.Errorf("Test failure")
	}
	return ioutil.ReadFile(fname)
}

func (s *Server) readSize() (int64, error) {
	var size int64
	err := filepath.Walk(s.path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func (s *Server) checkSize(ctx context.Context) error {
	if s.dirSize > 10*1024*1024 {
		s.RaiseIssue(ctx, "Lots of logging", fmt.Sprintf("There are %v logs - this is too much", s.dirSize), false)
	}
	return nil
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("logging", false, false)
	if err != nil {
		return
	}

	size, err := server.readSize()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	DirSize.Set(float64(size))
	server.dirSize = size

	server.RegisterRepeatingTaskNonMaster(server.checkSize, "check_size", time.Minute*5)

	fmt.Printf("%v", server.Serve())
}
