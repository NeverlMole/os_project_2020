package account

import (
  "fmt"

  pb "blockchaindb/protobuf/go"
  "blockchaindb/secure"
  "blockchaindb/util"
)

const UserIDLength = 8
const AccountInitValue = 1000

type AccountStates struct {
  data map[string]int32
}

func New() (*AccountStates) {
  return &AccountStates{make(map[string]int32)}
}

func CheckID(userID string) bool {
  if len(userID) != UserIDLength {
    return false
  }

  for _, char := range userID {
    if (char < 'a' || char > 'z') &&
       (char < 'A' || char > 'Z') &&
       (char < '0' || char > '9') {
      return false
    }
  }

  return true
}

func CheckServerID(serverID string) bool {
  if len(serverID) != UserIDLength {
    return false
  }

  if serverID[:6] != "Server" {
    return false
  }

  for _, char := range serverID[6:] {
    if char < '0' || char > '9' {
      return false
    }
  }
  return true
}

func (ac *AccountStates) Touch(userID string) {
  if _, found := ac.data[userID]; !found {
    ac.data[userID] = AccountInitValue
  }
}

func (ac *AccountStates) Get(userID string) (int32, error) {
  if !CheckID(userID) {
    return 0, secure.NewInvalidUserIDErr(userID)
  }

  ac.Touch(userID)

  return ac.data[userID], nil
}

func (ac *AccountStates) Apply(trans *pb.Transaction, serverID string) error {
  if err := ac.Check(trans, serverID); err != nil {
    return err
  }

  ac.data[trans.FromID] -= trans.Value
  ac.data[trans.ToID] += trans.Value - trans.MiningFee
  ac.data[serverID] += trans.MiningFee
  return nil
}

func (ac *AccountStates) Check(trans *pb.Transaction, serverID string) error {
  // Check the userID, fromID, toID.

  if !CheckID(trans.FromID) {
    return secure.NewInvalidUserIDErr(trans.FromID)
  }

  if !CheckID(trans.ToID) {
    return secure.NewInvalidUserIDErr(trans.ToID)
  }

  if !CheckServerID(serverID) {
    return secure.NewInvalidServerIDErr(serverID)
  }

  ac.Touch(trans.FromID)
  ac.Touch(trans.ToID)
  ac.Touch(serverID)

  if trans.MiningFee <= 0 || trans.Value <= trans.MiningFee {
    return secure.NewIntegrityConstrainErr(trans.MiningFee, trans.Value)
  }

  if ac.data[trans.FromID] < trans.Value {
    return secure.NewBalanceNotEnoughErr(trans.FromID, ac.data[trans.FromID],
                                           trans.Value)
  }

  return nil
}

func (ac *AccountStates) ApplyBlock(block *pb.Block) error {
  serverID := block.MinerID

  for _, trans := range block.Transactions {
    if err := ac.Apply(trans, serverID); err != nil {
      return secure.NewTransactionErr(trans.UUID, err)
    }
  }

  return nil
}

func (ac *AccountStates) ApplyBlocks(blocks []*pb.Block) error {
  for _, block := range blocks {
    if err := ac.ApplyBlock(block); err != nil {
      return secure.NewInvalidBlockErr(util.BlockHash(block), err)
    }
  }

  return nil
}

func (ac *AccountStates) String() string {
  return fmt.Sprintln(ac.data)
}

func (ac *AccountStates) Copy() (*AccountStates) {
  copyStates := make(map[string]int32)
  for key, v := range ac.data {
    copyStates[key] = v
  }

  return &AccountStates{data: copyStates}
}
