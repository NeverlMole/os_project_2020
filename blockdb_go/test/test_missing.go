package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
	pb "blockdb_go/protobuf/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)


func check(err error, message string) {
	if err != nil {
		log.Fatalf("Test RuntimeError:"+message+": %v", err)
		cmd.Process.Kill()
	}
}
func assert(exp int32, val int32, message string) {
	if val != exp {
		log.Fatalf("Test Failed: Expecting %s=%d, Got %d", message, exp, val)
		cmd.Process.Kill()
	}
}
func assertTrue(b bool, message string) {
	if !b {
		log.Fatalf("Test Failed: expect %s to be true", message)
		cmd.Process.Kill()
	}
}

func id(i int) string {
	return fmt.Sprintf("T1U%05d", i)
}


func clientGet(uid string) (int32, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return 0, err
  }
  c := pb.NewBlockDatabaseClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Get(ctx, &pb.GetRequest{UserID: uid})
  if err != nil {
    return 0, err
  }
  return r.Value, err
}

func clientDeposit(uid string, value int32) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return false, err
  }
  c := pb.NewBlockDatabaseClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Deposit(ctx, &pb.Request{UserID: uid,
                                       Value: value})
  if err != nil {
    return false, err
  }
  return r.Success, err
}

func clientPut(uid string, value int32) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return false, err
  }
  c := pb.NewBlockDatabaseClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Put(ctx, &pb.Request{UserID: uid,
                                   Value: value})
  if err != nil {
    return false, err
  }
  return r.Success, err
}


func clientTransfer(fid string, tid string, value int32) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return false, err
  }
  c := pb.NewBlockDatabaseClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Transfer(ctx, &pb.TransferRequest{FromID: fid, ToID: tid,
                                                Value: value})
  if err != nil {
    return false, err
  }
  return r.Success, err
}

func mustGet(uid string) int32 {
  for {
    value, err := clientGet(uid)
    if err == nil {
      return value
    }
  }
}

func mustDeposit(uid string, value int32) {
  for {
    result, _ := clientDeposit(uid, value)
    if result {
      return
    }
  }
}

func mustPut(uid string, value int32) {
  for {
    result, _ := clientPut(uid, value)
    if result {
      return
    }
  }
}

func checkServerStart() bool {
  test_id := "TESTTEST"
  _, err := clientGet(test_id)

  if err != nil {
    return false
  }

  return true
}

var cmd *exec.Cmd

var address, dataDir = func() (string, string) {
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

func main() {

  cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("rm %s*", dataDir))
  err := cmd.Start()

  cmd = exec.Command("./start.sh")
  err = cmd.Start()
  check(err, "start.sh")
  fmt.Println("Server start")
  time.Sleep(time.Millisecond * 500)

  num_operation := 103
  uid := "NONEXIST"

  for i := 1; i <= num_operation; i++ {
    mustPut(uid, 1)
  }

	err = cmd.Process.Kill()
	check(err, "Finished, kill server")
  fmt.Println("The server is killed.")

  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s1.json %s3.json", dataDir, dataDir))
  err = cmd.Start()
  cmd = exec.Command("./start.sh")
  err = cmd.Start()
  if checkServerStart() {
    log.Fatal("I don't know why the server started.")
  }
  fmt.Println("Good!")

  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s2.json %s4.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s3.json %s1.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("./start.sh")
  err = cmd.Start()
  if checkServerStart() {
    log.Fatal("I don't know why the server started.")
  }
  fmt.Println("Very good!")


  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s4.json %s2.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s1.json %s3.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("touch %s1.json", dataDir))
  cmd = exec.Command("./start.sh")
  err = cmd.Start()
  if checkServerStart() {
    log.Fatal("I don't know why the server started.")
  }
  fmt.Println("Almost succeed!")

  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s2.json %s4.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("mv %s3.json %s1.json", dataDir, dataDir))
  cmd.Start()
  cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("touch %s2.json", dataDir))
  cmd.Start()
  cmd = exec.Command("./start.sh")
  err = cmd.Start()
  if checkServerStart() {
    log.Fatal("I don't know why the server started.")
  }

	fmt.Println("Test con Passed.")
}
