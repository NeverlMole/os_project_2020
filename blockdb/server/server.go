package server

import (
	"golang.org/x/net/context"
	pb "blockdb/protobuf/go"
  log_pb "blockdb/log_protobuf/go"
  "blockdb/account"
)

var loglen int32

type Server struct {
  transNum int32
  orderNum int32
  curStates *account.AccountStates
}

func New(outputDir string, blockSize int) (*Server) {
  return &Server{curStates: account.New()}
}
// Database Interface
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
  s.transNum++

  value, err := s.curStates.Get(in.UserID)
	return &pb.GetResponse{Value: value}, err
}

func (s *Server) Apply(trans log_pb.Transaction) (bool, error) {
  return s.curStates.Apply(trans)
}

func (s *Server) Put(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	s.transNum++

  trans := log_pb.Transaction{Type: log_pb.Transaction_PUT,
                              UserID: in.UserID,
                              Value: in.Value,
                              TransID: s.transNum}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Deposit(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	s.transNum++

  trans := log_pb.Transaction{Type: log_pb.Transaction_DEPOSIT,
                              UserID: in.UserID,
                              Value: in.Value,
                              TransID: s.transNum}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Withdraw(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
	s.transNum++

  trans := log_pb.Transaction{Type: log_pb.Transaction_WITHDRAW,
                              UserID: in.UserID,
                              Value: in.Value,
                              TransID: s.transNum}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Transfer(ctx context.Context, in *pb.TransferRequest) (*pb.BooleanResponse, error) {
	s.transNum++

  trans := log_pb.Transaction{Type: log_pb.Transaction_TRANSFER,
                              FromID: in.FromID,
                              ToID: in.ToID,
                              Value: in.Value,
                              TransID: s.transNum}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

// Interface with test grader
func (s *Server) LogLength(ctx context.Context, in *pb.Null) (*pb.GetResponse, error) {
	return &pb.GetResponse{Value: s.transNum}, nil
}

