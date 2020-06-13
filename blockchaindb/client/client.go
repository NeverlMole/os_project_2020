package client

import (

	"golang.org/x/net/context"
	"google.golang.org/grpc"

  pb "blockchaindb/protobuf/go"
  "blockchaindb/util"
)

func ClientPushTransaction(serverList []string, transaction *pb.Transaction) {
	for _, address := range serverList{
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			continue
		}
		c := pb.NewBlockChainMinerClient(conn)
		defer conn.Close()

		ctx := context.Background()

		c.PushTransaction(ctx, transaction)
	}
}

func ClientPushBlock(serverList []string, block *pb.Block) {
	for _, address := range serverList{
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			continue
		}
		c := pb.NewBlockChainMinerClient(conn)
		defer conn.Close()

		ctx := context.Background()

    blockStr, _ := util.ProtobufToString(block)
    c.PushBlock(ctx, &pb.JsonBlockString{Json: blockStr})
	}
}

func ClientGetBlock(serverList []string, hash string) ([]*pb.Block) {
  var blockList []*pb.Block

  for _, address := range serverList{
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			continue
		}
		c := pb.NewBlockChainMinerClient(conn)
		defer conn.Close()

		ctx := context.Background()

    blockStr, err := c.GetBlock(ctx, &pb.GetBlockRequest{BlockHash: hash})

    if err != nil {
      continue
    }

    block := &pb.Block{}
    if err := util.StringToProtobuf(blockStr.Json, block); err != nil {
      continue
    }
	  blockList = append(blockList, block)
	}
	return blockList;
}

func ClientGetCurBlock(serverList []string) ([]*pb.Block) {
  var blockList []*pb.Block

  for _, address := range serverList{
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			continue
		}
		c := pb.NewBlockChainMinerClient(conn)
		defer conn.Close()

		ctx := context.Background()

    blockStr, err := c.GetCurBlock(ctx, &pb.Null{})

    if err != nil {
      continue
    }

    block := &pb.Block{}
    if err := util.StringToProtobuf(blockStr.Json, block); err != nil {
      continue
    }
	  blockList = append(blockList, block)
	}

  return blockList
}
