package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	pb "blockdb/protobuf/go"
  "blockdb/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const blockSize int32 = 50

// Main function, RPC server initialization
func main() {
	// Read config
	address, outputDir := func() (string, string) {
		conf, err := ioutil.ReadFile("config.json")
		if err != nil {
			panic(err)
		}
		var dat map[string]interface{}
		err = json.Unmarshal(conf, &dat)
		if err != nil {
			panic(err)
		}
		dat = dat["1"].(map[string]interface{}) // should be dat[myNum] in the future
		return fmt.Sprintf("%s:%s", dat["ip"], dat["port"]),
           fmt.Sprintf("%s",dat["dataDir"])
	}()

	// Bind to port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening: %s ...", address)

	// Create gRPC server
	s := grpc.NewServer()
	pb.RegisterBlockDatabaseServer(s, server.New(outputDir, blockSize))
	// Register reflection service on gRPC server.
	reflection.Register(s)
	// Start server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
