package blocks

import (
  "log"
  "fmt"

  pb "blockchaindb/protobuf/go"
  "blockchaindb/account"
  "blockchaindb/util"
  "blockchaindb/secure"
)

type BlockMap struct {
  outputDir string
  storedBlock map[string]*pb.Block
  storedStates map[string]*account.AccountStates
  logger *log.Logger
}

func NewBlockMap(outputDir string, logger *log.Logger) (*BlockMap) {
  return &BlockMap{outputDir: outputDir,
                   logger: logger,
                   storedBlock: make(map[string]*pb.Block),
                   storedStates: make(map[string]*account.AccountStates)}
}

func (bm *BlockMap) AddBlock(block *pb.Block) error {
  curHash := util.BlockHash(block)

  if _, found := bm.storedBlock[curHash]; found {
    return secure.NewAddBlockErr(fmt.Sprintf(
      "Block with hash [%s] already exists.", curHash))
  }

  var curStates *account.AccountStates

  if block.BlockID == 1 {
    curStates = account.New()
  } else {
    preHash := block.PrevHash
    if _, found := bm.storedStates[preHash]; !found {
      return secure.NewAddBlockErr("Previous States not exists.")
    }

    curStates = bm.storedStates[preHash].Copy()
  }

  if err := curStates.ApplyBlock(block); err != nil {
    return secure.NewAddBlockErr(fmt.Sprintf(
      "Transactions apply failed with error: %v", err))
  }

  bm.storedStates[curHash] = curStates
  bm.storedBlock[curHash] = block

  return nil
}

func (bm *BlockMap) GetBlock(blockHash string) (*pb.Block, error) {
  return bm.storedBlock[blockHash], nil
}

func (bm *BlockMap) GetBlockStates(blockHash string) (*account.AccountStates, error) {
  return bm.storedStates[blockHash], nil
}
