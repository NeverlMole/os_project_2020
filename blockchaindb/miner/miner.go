package miner

import (
  pb "blockchaindb/protobuf/go"
  "blockchaindb/client"
  "blockchaindb/hash"
)

type Miner struct {
  serverID string
  isMining bool
  miner_stop chan
  serverList []string
}

func New() (*Miner) {
}

func (*Miner) Start(transactions []*pb.Transaction, preBlock *pb.Block) {
}

func (*Miner) IsMining() bool {
}

func (*Miner) Stop() {
}

func (*Miner) Mining(transactions []*pb.Transaction, preBlock *pb.Block,
                     miner_stop chan) {
}
