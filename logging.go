package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	//DirSize - the print queue
	DirSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "logging_dirsize",
		Help: "The size of the logs",
	}, []string{"base"})
)

//Server main server type
type Server struct {
	*goserver.GoServer
	path    string
	dpath   string
	test    bool
	dirSize int64
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
		path:     "/media/scratch/logs/",
		dpath:    "/media/scratch/dlogs",
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
		&pbg.State{Key: "blah", Value: int64(12)},
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

func (s *Server) convert(line []string) *pb.Log {
	time, _ := time.Parse(time.RFC3339Nano, line[0])

	s.DLog(context.Background(), fmt.Sprintf("CONVERTING %v", line[4]))
	return &pb.Log{
		Timestamp: time.Unix(),
		Context:   line[2],
		Log:       line[4],
		Origin:    line[1],
	}
}

func (s *Server) loadDLog(fname, origin, ctx string) ([]*pb.Log, error) {
	if origin != "" && !strings.Contains(fname, origin) {
		return make([]*pb.Log, 0), nil
	}

	if s.test {
		return nil, fmt.Errorf("Built to fail for the test")
	}
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var logs []*pb.Log
	for scanner.Scan() {
		line := scanner.Text()
		elems := strings.Split(line, "|")

		if len(elems) == 3 && (ctx == "" || elems[1] == ctx) {
			logs = append(logs, s.convert(elems))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil

}

func (s *Server) clean() error {
	s.cleanAllLogs()
	size, err := s.readSize()
	if err != nil {
		return err
	}
	s.dirSize = size
	return nil
}

func (s *Server) readSize() (int64, error) {
	files, err := ioutil.ReadDir(s.path)
	if err != nil {
		return -1, err
	}

	dirSize := int64(0)
	for _, f := range files {
		if f.IsDir() {
			size, err := s.readDirSize(f.Name())
			if err != nil {
				return -1, err
			}
			DirSize.With(prometheus.Labels{"base": f.Name()}).Set(float64(size))
			dirSize += size
		}
	}
	return dirSize, nil
}

func (s *Server) readDirSize(base string) (int64, error) {
	var size int64
	err := filepath.Walk(s.path+base, func(_ string, info os.FileInfo, err error) error {
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
		s.RaiseIssue("Lots of logging", fmt.Sprintf("There are %v logs on %v - this is too much", s.dirSize, s.Registry.GetIdentifier()))
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

	err := server.RegisterServerV2("logging", false, true)
	if err != nil {
		return
	}

	size, err := server.readSize()
	server.dirSize = size

	server.clean()

	fmt.Printf("%v", server.Serve())
}
