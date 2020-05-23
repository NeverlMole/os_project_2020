package server

import (
	"golang.org/x/net/context"
  "log"

	pb "blockdb/protobuf/go"
  log_pb "blockdb/log_protobuf/go"
  "blockdb/account"
  "blockdb/util"
)

var loglen int32

type Server struct {
  transNum int32
  orderNum int32
  blockSize int32
  blockNum int32
  curStates *account.AccountStates
  curLog []*log_pb.Transaction
  logHelper *util.LogHelper
}

func New(outputDir string, blockSize int32) (*Server) {
  return &Server{curStates: account.New(),
                 blockSize: blockSize,
                 logHelper: util.NewLogHelper(blockSize, outputDir)}
}
// Database Interface
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
  value, err := s.curStates.Get(in.UserID)
	return &pb.GetResponse{Value: value}, err
}

func (s *Server) CheckFlushLog() error {
  if int32(len(s.curLog)) == s.blockSize {
    if err := s.logHelper.WriteBlock(s.blockNum + 1, s.curLog); err != nil {
      return err
    }
    if err := s.logHelper.CleanLog(); err != nil {
      return err
    }
    s.curLog = s.curLog[:0]
    s.blockNum++
  }

  return nil
}

func (s *Server) Apply(trans *log_pb.Transaction) (bool, error) {
  s.transNum++
  trans.TransID = s.transNum
  if err := s.curStates.Check(trans); err != nil {
    return false, err
  }

  // Apply check succeed and start to update the log.
  if err := s.CheckFlushLog(); err != nil {
    return false, err
  }

  trans.OrderID = s.orderNum + 1
  s.curLog = append(s.curLog, trans)

  if err := s.logHelper.WriteLog(int32(len(s.curLog)), trans); err != nil {
    return false, err
  }

  if err := s.curStates.Apply(trans); err != nil {
    log.Fatalf("Two same apply has different results: error: %v", err)
  }
  s.orderNum++
  return true, nil
}

func (s *Server) Put(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
  trans := &log_pb.Transaction{Type: log_pb.Transaction_PUT,
                              UserID: in.UserID,
                              Value: in.Value}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Deposit(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
  trans := &log_pb.Transaction{Type: log_pb.Transaction_DEPOSIT,
                              UserID: in.UserID,
                              Value: in.Value}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Withdraw(ctx context.Context, in *pb.Request) (*pb.BooleanResponse, error) {
  trans := &log_pb.Transaction{Type: log_pb.Transaction_WITHDRAW,
                              UserID: in.UserID,
                              Value: in.Value}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

func (s *Server) Transfer(ctx context.Context, in *pb.TransferRequest) (*pb.BooleanResponse, error) {
  trans := &log_pb.Transaction{Type: log_pb.Transaction_TRANSFER,
                              FromID: in.FromID,
                              ToID: in.ToID,
                              Value: in.Value}
  result, err := s.Apply(trans)
	return &pb.BooleanResponse{Success: result}, err
}

// Interface with test grader
func (s *Server) LogLength(ctx context.Context, in *pb.Null) (*pb.GetResponse, error) {
	return &pb.GetResponse{Value: int32(len(s.curLog))}, nil
}

