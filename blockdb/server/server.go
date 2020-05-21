package server

import (
	"golang.org/x/net/context"
	pb "blockdb/protobuf/go"
)

var data = make(map[string]int32)
var loglen int32

type Server struct{}

func New(outputDir string, blockSize int) (*Server) {
  return &Server{}
}
// Database Interface
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	return &pb.GetResponse{Value: data[in.UserID]}, nil
}
func (s *Server) Put(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	loglen++
	data[in.UserID] = in.Value
	return &pb.BooleanResponse{Success: true}, nil
}
func (s *Server) Deposit(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	loglen++
	data[in.UserID] += in.Value
	return &pb.BooleanResponse{Success: true}, nil
}
func (s *Server) Withdraw(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	loglen++
	data[in.UserID] -= in.Value
	return &pb.BooleanResponse{Success: true}, nil
}
func (s *Server) Transfer(ctx context.Context, in *pb.TransferRequest) (*pb.BooleanResponse, error) {
	loglen++
	data[in.FromID] -= in.Value
	data[in.ToID] += in.Value
	return &pb.BooleanResponse{Success: true}, nil
}
// Interface with test grader
func (s *Server) LogLength(ctx context.Context, in *pb.Null) (*pb.GetResponse, error) {
	return &pb.GetResponse{Value: loglen}, nil
}

