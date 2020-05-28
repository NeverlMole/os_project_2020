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

  cmd := exec.Command("./start.sh")
  err := cmd.Start()
  check(err, "start.sh")
  fmt.Println("Server start")
  time.Sleep(time.Millisecond * 500)

  c_stop := make(chan int)
  num_clients := 20
  salary := 1000
  emp_id := "poolpony"
  game_id := "gamegame"
  months := 100

	var wg sync.WaitGroup

  for i := 1; i <= num_clients; i++ {
    wg.Add(1)
    go func(c chan int, i int) {
      for {
        result, _ := clientTransfer(emp_id, game_id, int32(i))
        if result {
          fmt.Printf("Game %d let %s waste %d cainn.\n", i, emp_id, i)
        }
        time.Sleep(time.Millisecond * 100)

        select {
        case <-c:
          wg.Done()
          return
        default:
          continue
        }
      }
    }(c_stop, i)
  }

  mustPut(emp_id, 0)
  mustPut(game_id, 0)
  fmt.Printf("%s had no money, but he found a job.\n", emp_id)

  for i := 1; i <= months; i++ {
    mustDeposit(emp_id, int32(salary))
    fmt.Printf("Great! %s got %d cainn in month %d.\n", emp_id, salary, i)

    for {
      value := mustGet(emp_id)
      if value == 0 {
        value_game := mustGet(game_id)
        assert(int32(i * salary), value_game, "Verify the balance")
        fmt.Printf("Poor %s! He has wasted all his money (%d) on game!\n",
                   emp_id, value_game)
        break
      }
	    time.Sleep(time.Millisecond * 10)
    }
  }

  fmt.Printf("Now %s lost his job because he spent too much time on game!\n",
             emp_id)

  for i := 1; i <= num_clients; i++ {
    c_stop <- 0
  }
  wg.Wait()
	err = cmd.Process.Kill()
	check(err, "Finished, kill server")
	fmt.Println("Test con Passed.")
}
