package server

import (
	"golang.org/x/net/context"
  "log"

	pb "blockdb_go/protobuf/go"
  log_pb "blockdb_go/log_protobuf/go"
  "blockdb_go/account"
  "blockdb_go/util"
  "blockdb_go/secure"
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

func (s *Server) Init() error {
  var lastBlockErr error = nil
  testStates := account.New()

  // Read the blocks
  for ; ;s.blockNum++ {
    logs, err := s.logHelper.ReadBlock(s.blockNum + 1)
    if err != nil {
      lastBlockErr = err
      break
    }

    if logs == nil {
      break
    }

    if err := testStates.ApplyLog(logs); err != nil {
      lastBlockErr = secure.NewBlockInvalidErr(s.blockNum + 1)
      break
    }
    if err := s.curStates.ApplyLog(logs); err != nil {
      log.Panic(secure.NewServerInitErr(secure.NewApplyInconsistentErr()))
    }
  }

  // Read the log
  logs, err := s.logHelper.ReadLogs()
  if err != nil {
    log.Panic(secure.NewServerInitErr(err))
  }
  if lastBlockErr != nil && int32(len(logs)) != s.blockSize {
    log.Panic(secure.NewServerInitErr(lastBlockErr))
  }

  if len(logs) != 0 {
    if s.blockSize * s.blockNum + 1 < logs[0].OrderID {
      log.Panic(secure.NewServerInitErr(secure.NewDataMissingErr("blocks")))
    }

    if s.blockSize * s.blockNum + 1 > logs[0].OrderID {
      // The logs were stored in blocks
      logs = logs[:0]
    }
  }

  if err := testStates.ApplyLog(logs); err != nil {
    log.Panic(secure.NewServerInitErr(err))
  }
  if err := s.curStates.ApplyLog(logs); err != nil {
    log.Panic(secure.NewServerInitErr(secure.NewApplyInconsistentErr()))
  }
  s.curLog = logs
  s.orderNum = s.blockSize * s.blockNum + int32(len(s.curLog))

  log.Printf("Init complete with %d block, %d logs and account states:\n",
              s.blockNum, len(s.curLog))
  s.curStates.Print()

  return nil
}

// Flush the log if the length is equal to the blockSize
func (s *Server) CheckFlushLog() error {
  if int32(len(s.curLog)) == s.blockSize {
    if err := s.logHelper.WriteBlock(s.blockNum + 1, s.curLog); err != nil {
      return err
    }
    if err := s.logHelper.CleanLogs(); err != nil {
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
    return false, secure.NewTransactionErr(trans.TransID, err)
  }

  // Apply check succeed and start to update the log.
  if err := s.CheckFlushLog(); err != nil {
    return false, secure.NewTransactionErr(trans.TransID, err)
  }

  trans.OrderID = s.orderNum + 1

  if err := s.logHelper.WriteLog(int32(len(s.curLog)) + 1, trans); err != nil {
    return false, secure.NewTransactionErr(trans.TransID, err)
  }

  if err := s.curStates.Apply(trans); err != nil {
    log.Panic(secure.NewApplyInconsistentErr())
  }
  s.orderNum++
  s.curLog = append(s.curLog, trans)
  return true, nil
}

/************************-Database Interface-*****************************/
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
  value, err := s.curStates.Get(in.UserID)
  if err != nil {
    err = secure.NewUserGetErr(err)
  }
  return &pb.GetResponse{Value: value}, err
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

