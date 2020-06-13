package miner

import (
  "fmt"
  "log"

  pb "blockchaindb/protobuf/go"
  "blockchaindb/client"
  "blockchaindb/hash"
  "blockchaindb/util"
)

const MAX_NONCE int32 = 99999999

type Miner struct {
  serverID string
  isMining bool
  minerStop chan bool
  serverList []string
  logger *log.Logger
}

func New(serverID string, serverList []string, logger *log.Logger) (*Miner) {
  return &Miner{serverID : serverID,
                isMining : false,
                serverList: serverList,
                logger: logger}
}

func (m *Miner) Start(transactions []*pb.Transaction, preBlock *pb.Block) {
	if m.isMining {
    m.logger.Fatalf("Miner starts misteriously!")
	}

  m.logger.Printf("Miner starts mining.\n")

  // Create a block
  transSlice := make([]*pb.Transaction, len(transactions))
  copy(transSlice, transactions)
  block := &pb.Block{Transactions: transSlice, MinerID: m.serverID}

  if preBlock != nil {
	  block.BlockID = preBlock.BlockID + 1
	  block.PrevHash = util.BlockHash(preBlock)
  } else {
	  block.BlockID = 1
	  block.PrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
  }

	m.minerStop = make(chan bool, 1)
	m.isMining = true
	go m.Mining(block, m.minerStop)
}

func (m *Miner) IsMining() bool {
	return m.isMining
}

func (m *Miner) Stop() {
	if !m.isMining {
    return
  }
  m.logger.Printf("Miner stops mining.\n")

  m.minerStop <- true
  m.isMining = false
}

func (m *Miner) Mining(block *pb.Block, minerStop chan bool) {
  for i := 0; int32(i) <= MAX_NONCE; i++ {
    block.Nonce = fmt.Sprintf("%08d", i)
    if hash.CheckHash(util.BlockHash(block)) {
	    client.ClientPushBlock(m.serverList, block)
      return
		}

    // Stop if message from minerStop
    select {
    case <-minerStop:
      return
    default:
      continue
    }
	}
}
