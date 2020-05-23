package util

import (
  "os"
  "strconv"
  "log"

  "github.com/golang/protobuf/jsonpb"
  "github.com/golang/protobuf/proto"

  log_pb "blockdb/log_protobuf/go"
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

func LogFileName(num int32) string {
  return strconv.Itoa(int(num)) + ".log.json"
}

func BlockFileName(num int32) string {
  return strconv.Itoa(int(num)) + ".json"
}

func (lh *LogHelper) WriteLog(num int32, trans *log_pb.Transaction) error {
  err := ProtobufWrite(lh.outputDir + LogFileName(num), trans)

  if err != nil {
    // TODO: wrap the error by sentence like "Write log fail:".
    return err
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
    // TODO: wrap the error by sentence like "Write log fail:".
    return err
  }

  return nil
}

func (lh *LogHelper) CleanLog() error {
  for i := lh.blockSize; i > 0; i-- {
    if PathIsExist(lh.outputDir + LogFileName(i)) {
      if err := os.Remove(lh.outputDir + LogFileName(i)); err != nil {
        log.Fatalf("Removing log file failed: %v\n", err)
      }
    }
  }
  return nil
}
