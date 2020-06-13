package util

import (
  "github.com/golang/protobuf/jsonpb"
  "github.com/golang/protobuf/proto"

  pb "blockchaindb/protobuf/go"
  "blockchaindb/hash"
)

func ProtobufToString(msg proto.Message) (string, error) {
  m := jsonpb.Marshaler{}
  return m.MarshalToString(msg)
}

func StringToProtobuf(s string, msg proto.Message) error {
  if err := jsonpb.UnmarshalString(s, msg); err != nil {
    return err
  }

  return nil
}

func BlockIsBetter(block *pb.Block, cblock *pb.Block) bool {
  if block == nil {
    return false
  }

  if cblock == nil {
    return true
  }

  if block.BlockID > cblock.BlockID {
    return true
  }

  if block.BlockID < cblock.BlockID {
    return false
  }

  if len(BlockHash(block)) < len(BlockHash(cblock)) {
    return true
  }
  return false
}

func BlockHash(block *pb.Block) string {
  blockStr, _ := ProtobufToString(block)
  return hash.GetHashString(blockStr)
}
