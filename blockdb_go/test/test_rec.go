package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
  "sync"
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
	// Set up a connection to the server.

//  cmd = exec.Command(fmt.Sprintf("ls"))
 // err := cmd.Start()
 // check(err, "clean data")

  c_stop := make(chan int)

	var wg sync.WaitGroup
  wg.Add(1)
  go func(c chan int) {
    for {
      cmd := exec.Command("./start.sh")
	    err := cmd.Start()
	    check(err, "start.sh")
      fmt.Println("Server start")
	    time.Sleep(time.Millisecond * 500)

	    err = cmd.Process.Kill()
	    check(err, "Finished, kill server")
      fmt.Println("Server crash")

      select {
      case <-c:
        wg.Done()
        return
      default:
        continue
      }
    }
  }(c_stop)

  uid := "NONEXIST"

  init_value := mustGet(uid)

  for i := 1; i <= 400; i++ {
    mustDeposit(uid, 1)

    value := mustGet(uid)
	  assert(init_value + int32(i), value, "Verify the balance")
    fmt.Printf("Deposit %d succeed\n", i)
	  time.Sleep(time.Millisecond * 10)
  }

  c_stop <- 0
  wg.Wait()
	fmt.Println("Test rec Passed.")
}
