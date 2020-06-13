package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
  "sync"
	"os/exec"
	"time"
	pb "blockchaindb/protobuf/go"
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


func clientGet(address string, uid string) (int32, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return 0, err
  }
  c := pb.NewBlockChainClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Get(ctx, &pb.GetRequest{UserID: uid})
  if err != nil {
    return 0, err
  }
  return r.Value, err
}

func clientTransfer(address string, fid string, tid string, value int32, mfee int32, uid string) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
    return false, err
  }
  c := pb.NewBlockChainMinerClient(conn)
	defer conn.Close()

	ctx := context.Background()

  r, err := c.Transfer(ctx, &pb.TransferRequest{FromID: fid, ToID: tid,
                                                Value: value, Miningfee: mfee, UUID = uid})
  if err != nil {
    return false, err
  }
  return r.Success, err
}

func mustallGet(uid string) []int32 {
  valueList = make([]int32, 0)
  for _, address := range serverList{
    for {
      value, err := clientGet(address, uid)
      if err == nil {
        valueList = append(valueList, value)
        break
      }
    }
  }
}

func mustallTransfer(fid string, tid string, value int32, mfee int32, uid string) {
  for _, address := range serverList{
    for {
      result, _ := clientTransfer(address, fid, tid, value, mfee, uid)
      if result {
        break
      }
    }
  }
}

func UUID128bit() string {
	// Returns a 128bit hex string, RFC4122-compliant UUIDv4
	u:=make([]byte,16)
	_,_=rand.Read(u)
	// this make sure that the 13th character is "4"
	u[6] = (u[6] | 0x40) & 0x4F
	// this make sure that the 17th is "8", "9", "a", or "b"
	u[8] = (u[8] | 0x80) & 0xBF 
	return fmt.Sprintf("%x",u)
}

var cmd *exec.Cmd

func get_address(string id) string {
  conf, err := ioutil.ReadFile("config.json")
  if err != nil {
    panic(err)
  }
  var dat map[string]interface{}
  err = json.Unmarshal(conf, &dat)
  if err != nil {
    panic(err)
  }
  dat = dat[id].(map[string]interface{}) 
  return fmt.Sprintf("%s:%s", dat["ip"], dat["port"])
}

var serverList := make([]string, 0)

func main() {

  var wg sync.WaitGroup
  wg.Add(1)

  server_num := 20

  for i := 1; i <= server_num; i++ {
    cmd := exec.Command("./start.sh --id=" + fmt.Sprintf("%s",i));
    err := cmd.Start()
    check(err, "start.sh")
    fmt.Println("Server start")
    time.Sleep(time.Millisecond * 500)
    serverList = append(serverList, get_address(fmt.Sprintf("%s",i)))
  }
 
  uid1 := "NONEXIST"
  uid2 := "MOMEXIST"

  init_value1 := mustallGet(uid1)
  init_value2 := mustallGet(uid2)
 for j := 1; j <= server_num; j++{
  	assert(init_value1[j], init_value1[1], "Verify the balance")
    assert(init_value2[j], init_value2[1], "Verify the balance")
  }

  for i := 1; i <= 400; i++ {
    mustalltransfer(uid1, uid2, 1, 0, UUID128bit())

    value1 := mustallGet(uid1)
    value2 := mustallGet(uid2)
    for j := 1; j <= server_num; j++{
  	  assert(init_value1[j] - int32(i), value1[1], "Verify the balance")
      assert(init_value2[j] + int32(i), value2[1], "Verify the balance")
    }

    fmt.Printf("transfer %d succeed\n", i)
	  time.Sleep(time.Millisecond * 10)
  }

  wg.Wait()
  fmt.Println("Test Consistency Passed.")
}
