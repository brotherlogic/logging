package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/logging/proto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"

	pbg "github.com/brotherlogic/goserver/proto"
)

func init() {
	resolver.Register(&utils.DiscoveryServerResolverBuilder{})
}

//Server main server type
type Server struct {
	*goserver.GoServer
	path    string
	dirSize int64
	test    bool
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
	server.dirSize = size
	fmt.Printf("%v", server.Serve())
}
