package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
  "flag"
  "os"

	pb "blockchaindb/protobuf/go"
  "blockchaindb/server"
  "blockchaindb/util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const blockSize int32 = 50
const pendingLen int32 = 6

var id = flag.Int("id",1,"Server's ID, 1<=ID<=NServers")

// Main function, RPC server initialization
func main() {
  // Get ID
  flag.Parse()
	serverID := fmt.Sprintf("%d", *id)

  // Read config
  conf, err := ioutil.ReadFile("config.json")
  if err != nil {
    panic(err)
  }
  var dat map[string]interface{}
  err = json.Unmarshal(conf, &dat)
  if err != nil {
    panic(err)
  }

  if _, ok := dat[serverID]; !ok {
    panic(fmt.Sprintf("Server with id %s is not found in the config file",
          serverID))
  }
  server_dat := dat[serverID].(map[string]interface{})
  address := fmt.Sprintf("%s:%s", server_dat["ip"], server_dat["port"])
  outputDir := fmt.Sprintf("%s", server_dat["dataDir"])

  // Get Server List
  serverList := make(map[string]string)
  for id, _ := range dat {
    if id == "nservers" {
      continue
    }
    server_dat := dat[id].(map[string]interface{})
    serverList[id] = fmt.Sprintf("%s:%s", server_dat["ip"], server_dat["port"])
  }

  // Create the logger for the server
  outputDir = util.Directorize(outputDir)
  if !util.PathIsExist(outputDir) {
    panic(fmt.Sprintf("The outputDir %s not exists.", outputDir))
  }
  f, err := os.OpenFile(outputDir + "log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
	  log.Println(err)
  }
  defer f.Close()

  logger := log.New(f, "prefix", log.LstdFlags)
  logger.Printf("Server %s is starting ...\n", serverID)

	// Bind to port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatalf("failed to listen: %v\n", err)
	}
	logger.Printf("Listening: %s ...\n", address)

	// Create gRPC server
	s := grpc.NewServer()
  dbserver := server.New(int32(*id), outputDir, blockSize, pendingLen,
                         serverList, logger)

  if err := dbserver.Init(); err != nil {
    logger.Fatalf("Init failed with err: %v", err)
  }

	pb.RegisterBlockChainMinerServer(s, dbserver)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	// Start server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
