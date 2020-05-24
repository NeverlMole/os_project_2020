package util

import (
  "os"
  "strconv"
  "log"

  "github.com/golang/protobuf/jsonpb"
  "github.com/golang/protobuf/proto"

  log_pb "blockdb_go/log_protobuf/go"
  "blockdb_go/secure"
)

type LogHelper struct {
  blockSize int32
  outputDir string
}

func NewLogHelper(blockSize int32, outputDir string) (*LogHelper) {
  outputDir = Directorize(outputDir)
  if !PathIsExist(outputDir) {
    log.Fatalf("The output directory " + outputDir + " does not exist.")
  }
  return &LogHelper{blockSize, outputDir}
}

func ProtobufWrite(dir string, msg proto.Message) error {
  m := jsonpb.Marshaler{}
  fo, err := os.Create(dir)
  if err != nil {
    return err
  }

  if err := m.Marshal(fo, msg); err != nil {
    return err
  }

  if err := fo.Close(); err != nil {
    return err
  }

  return nil
}

func ProtobufToString(msg proto.Message) (string, error) {
  m := jsonpb.Marshaler{}
  return m.MarshalToString(msg)
}

func ProtobufRead(dir string, msg proto.Message) error {
  fo, err := os.Open(dir)
  if err != nil {
    return err
  }

  if err := jsonpb.Unmarshal(fo, msg); err != nil {
    return err
  }

  if err := fo.Close(); err != nil {
    return err
  }

  return nil
}


func LogFileName(num int32) string {
  return strconv.Itoa(int(num)) + ".log.json"
}

func BlockFileName(num int32) string {
  return strconv.Itoa(int(num)) + ".json"
}

func (lh *LogHelper) WriteLog(num int32, trans *log_pb.Transaction) error {
  err := ProtobufWrite(lh.outputDir + LogFileName(num), trans)

  if err != nil {
    return secure.NewWriteLogFailErr(num, err)
  }

  return nil
}

func (lh *LogHelper) WriteBlock(num int32, log_slice []*log_pb.Transaction) error {
  if int32(len(log_slice)) != lh.blockSize {
    log.Fatalf("Try to create a block with wrong size (%d)\n", len(log_slice))
  }
  err := ProtobufWrite(lh.outputDir + BlockFileName(num),
                       &log_pb.Block{BlockID: num,
                                     PrevHash: "00000000",
                                     Transactions: log_slice,
                                     Nonce: "00000000"})
  if err != nil {
    return secure.NewWriteBlockFailErr(num, err)
  }

  return nil
}

func (lh *LogHelper) CleanLogs() error {
  for i := lh.blockSize; i > 0; i-- {
    if PathIsExist(lh.outputDir + LogFileName(i)) {
      if err := os.Remove(lh.outputDir + LogFileName(i)); err != nil {
        log.Fatalf("Removing log file failed: %v\n", err)
      }
    }
  }
  return nil
}

func (lh *LogHelper) ReadBlock(num int32) ([]*log_pb.Transaction, error) {
  if !PathIsExist(lh.outputDir + BlockFileName(num)) {
    // There is no block with num id.
    return nil, nil
  }

  block := &log_pb.Block{}

  if err := ProtobufRead(lh.outputDir + BlockFileName(num), block);
     err != nil {
    return nil, secure.NewBlockInvalidErr(num)
  }

  if err := lh.CheckBlock(num, block); err != nil {
    return nil, err
  }

  return block.Transactions, nil
}

func (lh *LogHelper) CheckBlock(num int32, block *log_pb.Block) error {
  if (int32(len(block.Transactions)) != lh.blockSize) ||
     (block.BlockID != num) {
    return secure.NewBlockInvalidErr(num)
  }

  for i, trans := range block.Transactions {
    if trans.OrderID != (num - 1) * lh.blockSize + int32(i) + 1 {
      return secure.NewBlockInvalidErr(num)
    }
  }
  return nil
}

func (lh *LogHelper) ReadLogs() ([]*log_pb.Transaction, error) {
  preMissing := false
  logs := []*log_pb.Transaction{}

  for i := int32(1); i <= lh.blockSize; i++ {
    if !PathIsExist(lh.outputDir + LogFileName(i)) {
      preMissing = true
      continue
    }

    if preMissing == true {
      return nil, secure.NewDataMissingErr("logs")
    }

    trans := &log_pb.Transaction{}
    if err := ProtobufRead(lh.outputDir + LogFileName(i), trans); err != nil {
      preMissing = true
      continue
    }

    if i > 1 && trans.OrderID != logs[i-2].OrderID + 1 {
      preMissing = true
      continue
    }

    logs = append(logs, trans)
  }

  return logs, nil
}
