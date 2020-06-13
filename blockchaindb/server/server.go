package server

import (
	"golang.org/x/net/context"
  "log"
  "sync"
  "fmt"

	pb "blockchaindb/protobuf/go"
  "blockchaindb/account"
  "blockchaindb/util"
  "blockchaindb/secure"
  "blockchaindb/hash"
  "blockchaindb/blocks"
  "blockchaindb/miner"
  "blockchaindb/client"
)

var loglen int32

type Server struct {
  serverID int32
  serverUserID string
  BlockSize int32
  PendingBlockLength int32
  mux sync.Mutex
  curBlock *pb.Block
  curStates *account.AccountStates
  confirmedStates *account.AccountStates
  blockMap *blocks.BlockMap
  pendingTrans []*pb.Transaction
  miner *miner.Miner
  serverList []string
  logger *log.Logger
}

func New(serverID int32, outputDir string, blockSize int32, pendingLen int32,
         serverList map[string]string, logger *log.Logger) (*Server) {
  IDstr := fmt.Sprintf("%d", serverID)
  serverUserID := fmt.Sprintf("Server%02d", serverID)

  var noSelfServerList, allServerList []string = nil, nil
  for id, address := range serverList {
    allServerList = append(allServerList, address)
    if id != IDstr {
      noSelfServerList = append(noSelfServerList, address)
    }
  }

  return &Server{serverID: serverID,
                 serverUserID: serverUserID,
                 BlockSize: blockSize,
                 PendingBlockLength: pendingLen,
                 blockMap: blocks.NewBlockMap(outputDir, logger),
                 miner: miner.New(serverUserID, allServerList, logger),
                 serverList: noSelfServerList,
                 logger: logger}
}

func (s *Server) CurLength() int32 {
  if s.curBlock != nil {
    return s.curBlock.BlockID
  }
  return 0
}

func (s *Server) Init() error {
  s.logger.Println("Server start init.")

  var bestBlock *pb.Block = nil
  curBlocks := client.ClientGetCurBlock(s.serverList)

  for _, block := range curBlocks {
    r, _ := s.UpdateBranch(block)
    if !r {
      continue;
    }

    if util.BlockIsBetter(block, bestBlock) {
      bestBlock = block
    }
  }

  if bestBlock == nil {
    // Clean start
    s.curBlock = nil
    s.pendingTrans = nil
    s.curStates = account.New()
    s.confirmedStates = account.New()
  } else {
    err := s.InitBranch(bestBlock)
    if err!= nil {
      return err
    }
  }

  s.logger.Printf("Init complete with current branch length %d,\n",
                  s.CurLength())
  s.logger.Printf("with current states: %s\n", s.curStates.String())
  s.logger.Printf("and confirmed states: %s.\n", s.confirmedStates.String())

  return nil
}

func (s *Server) InitBranch(block *pb.Block) error {
  var err error

  s.miner.Stop()
  s.curBlock = block
  s.pendingTrans = nil
  s.curStates, err = s.blockMap.GetBlockStates(util.BlockHash(block))
  if err != nil {
    return err
  }

  if block.BlockID <= s.PendingBlockLength {
    s.confirmedStates = account.New()
    return nil
  }

  preBlock := block
  for i := 1; int32(i) < s.PendingBlockLength; i++ {
    preBlock, _ = s.blockMap.GetBlock(preBlock.PrevHash)
    if preBlock == nil {
      return secure.NewBlockMapBrokenErr()
    }
  }

  s.confirmedStates, err = s.blockMap.GetBlockStates(preBlock.PrevHash)
  if err != nil {
    return err
  }

  return nil
}

func (s *Server) UpdateBranch(block *pb.Block) (bool, error) {
  s.logger.Printf("Try to update current branch.\n")

  curHash := util.BlockHash(block)

  if !hash.CheckHash(curHash) {
    s.logger.Printf("Update failed because invalid hash.")
    return false, nil
  }

  if r, _ := s.blockMap.GetBlock(util.BlockHash(block)); r != nil {
    s.logger.Printf("Update failed because the block is already added.\n")
    return false, nil
  }

  var branch []*pb.Block
  branchLen := block.BlockID
  preBlock := block
  var preStates *account.AccountStates = nil
  branch = append(branch, block)

  for i := branchLen; i > 1; i-- {
    preHash := preBlock.PrevHash

    if !hash.CheckHash(preHash) {
      s.logger.Printf("Update failed because cannot get a complete branch.\n")
      return false, nil
    }

    if r, _ := s.blockMap.GetBlockStates(preHash); r!= nil {
      preStates = r
      break
    }

    // Find an valid block
    posBlocks := client.ClientGetBlock(s.serverList, preHash)
    validFound := false
    for _, b := range posBlocks {
      if (b.BlockID != preBlock.BlockID - 1) ||
         (util.BlockHash(b) != preHash) {
         continue
      }
      validFound = true
      preBlock = b
      break
    }

    if validFound {
      branch = append(branch, preBlock)
    } else {
      s.logger.Printf("Update failed because cannot get a complete branch.\n")
      return false, nil
    }
  }

  if preStates == nil {
    preStates = account.New()
  }

  // Reverse the branch
  curLen := len(branch)
  branch_r := make([]*pb.Block, curLen)
  for i, block := range branch {
    branch_r[curLen - i - 1] = block
  }

  if err := preStates.ApplyBlocks(branch_r); err != nil {
    return false, err
  }

  for _, b := range branch_r {
    if err := s.blockMap.AddBlock(b); err != nil {
      s.logger.Fatalf("Error occurs when adding a block: %v\n", err)
    }
  }

  s.logger.Printf("Added a new branch ends by block with hash [%s].\n",
                  util.BlockHash(block))

  return true, nil
}

func (s *Server) Apply(trans *pb.Transaction) (bool, error) {
  // Integrity Constraints
  if trans.MiningFee <= 0 || trans.Value <= trans.MiningFee {
    return false, nil
  }

  if err := s.curStates.Apply(trans, s.serverUserID); err != nil {
    s.logger.Printf("The transaction %s failed with error: %v\n", trans.UUID,
                    err)
    return true, secure.NewTransactionErr(trans.UUID, err)
  }

  s.pendingTrans = append(s.pendingTrans, trans)
  s.TryMining()

  s.logger.Printf("The transaction %s is now pendding.\n", trans.UUID)
  return true, nil
}

func (s *Server) TryMining() {
  if int32(len(s.pendingTrans)) > s.BlockSize {
    return
  }
  s.miner.Stop();
  s.miner.Start(s.pendingTrans, s.curBlock)
}

/************************-Database Interface-*****************************/
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  value, _ := s.confirmedStates.Get(in.UserID)
  return &pb.GetResponse{Value: value}, nil
}

func (s *Server) Transfer(ctx context.Context, trans *pb.Transaction) (*pb.BooleanResponse, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  transStr, _ := util.ProtobufToString(trans)
  s.logger.Printf("Received a new transaction from client: %s\n", transStr)

  client.ClientPushTransaction(s.serverList, trans)

  result, err := s.Apply(trans)
  if !result {
    s.logger.Printf("The transaction %s failed with error: %v\n", trans.UUID,
      err)
  }

	return &pb.BooleanResponse{Success: result}, nil
}

func (s *Server) Verify(ctx context.Context, trans *pb.Transaction) (*pb.VerifyResponse, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  // search in pendingTrans
  for _, t := range s.pendingTrans {
    if t.UUID == trans.UUID {
      return &pb.VerifyResponse{Result: pb.VerifyResponse_PENDING}, nil
    }
  }


  if s.curBlock == nil {
    return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED}, nil
  }
  preBlock := s.curBlock
  for i := 1; ; i++ {
    for _, t := range preBlock.Transactions {
      if t.UUID == trans.UUID {
        if int32(i) > s.PendingBlockLength {
          return &pb.VerifyResponse{Result: pb.VerifyResponse_SUCCEEDED}, nil
        } else {
          return &pb.VerifyResponse{Result: pb.VerifyResponse_PENDING}, nil
        }
      }
    }

    if preBlock.BlockID <= 1 {
      break
    }
    preBlock, _ = s.blockMap.GetBlock(preBlock.PrevHash)
    if preBlock == nil {
      s.logger.Fatalf("%v", secure.NewBlockMapBrokenErr())
    }
  }

  return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED}, nil
}

func (s *Server) GetHeight(ctx context.Context, in *pb.Null) (*pb.GetHeightResponse, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

	return &pb.GetHeightResponse{Height: s.CurLength()}, nil
}

func (s *Server) GetBlock(ctx context.Context, in *pb.GetBlockRequest) (*pb.JsonBlockString, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  block, err := s.blockMap.GetBlock(in.BlockHash)

  if block == nil || err != nil {
    return &pb.JsonBlockString{Json: "Null"}, nil
  }

  blockStr, _ := util.ProtobufToString(block)
	return &pb.JsonBlockString{Json: blockStr}, nil
}

func (s *Server) GetCurBlock(ctx context.Context, in *pb.Null) (*pb.JsonBlockString, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  if s.curBlock == nil {
    return &pb.JsonBlockString{Json: "Null"}, nil
  }

  blockStr, _ := util.ProtobufToString(s.curBlock)
	return &pb.JsonBlockString{Json: blockStr}, nil
}

func (s *Server) PushTransaction(ctx context.Context, trans *pb.Transaction) (*pb.Null, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  transStr, _ := util.ProtobufToString(trans)
  s.logger.Printf("Received a new transaction: %s\n", transStr)

  s.Apply(trans)

	return &pb.Null{}, nil
}


func (s *Server) PushBlock(ctx context.Context, blockStr *pb.JsonBlockString) (*pb.Null, error) {
  s.mux.Lock()
  defer s.mux.Unlock()

  s.logger.Printf("Received a new block: %s\n", blockStr.Json)

  var block *pb.Block = &pb.Block{}
  if err := util.StringToProtobuf(blockStr.Json, block); err != nil {
    return &pb.Null{}, nil
  }

  if !util.BlockIsBetter(block, s.curBlock) {
    return &pb.Null{}, nil
  }

  if r, _ := s.UpdateBranch(block); r {
    if err := s.InitBranch(block); err != nil {
      s.logger.Fatalf("Cannot init after update the branch with error: %v",
        err)
    }
  }

	return &pb.Null{}, nil
}
