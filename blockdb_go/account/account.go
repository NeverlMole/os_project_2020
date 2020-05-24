package account

import (
  "log"

  log_pb "blockdb_go/log_protobuf/go"
  "blockdb_go/secure"
)

const UserIDLength = 8

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

func (ac *AccountStates) Get(userID string) (int32, error) {
  if !CheckID(userID) {
    return 0, secure.NewInvalidUserIDErr(userID)
  }
  return ac.data[userID], nil
}

func (ac *AccountStates) Apply(trans *log_pb.Transaction) error {
  if err := ac.Check(trans); err != nil {
    return err
  }

  switch trans.Type {
  case log_pb.Transaction_PUT:
    ac.data[trans.UserID] = trans.Value
  case log_pb.Transaction_DEPOSIT:
    ac.data[trans.UserID] += trans.Value
  case log_pb.Transaction_WITHDRAW:
    ac.data[trans.UserID] -= trans.Value
  case log_pb.Transaction_TRANSFER:
    ac.data[trans.FromID] -= trans.Value
    ac.data[trans.ToID] += trans.Value
  }

  return nil
}

func (ac *AccountStates) ApplyLog(logs []*log_pb.Transaction) error {
  for i, trans := range logs {
    if err := ac.Apply(trans); err != nil {
      return secure.NewLogsInvalidErr(int32(i + 1), err)
    }
  }
  return nil
}

func (ac *AccountStates) Check(trans *log_pb.Transaction) error {
  // Check the userID, fromID, toID.
  switch trans.Type {
  case log_pb.Transaction_PUT, log_pb.Transaction_DEPOSIT,
       log_pb.Transaction_WITHDRAW:
    if !CheckID(trans.UserID) {
      return secure.NewInvalidUserIDErr(trans.UserID)
    }
  case log_pb.Transaction_TRANSFER:
    if !CheckID(trans.FromID) {
      return secure.NewInvalidUserIDErr(trans.FromID)
    }
    if !CheckID(trans.ToID) {
      return secure.NewInvalidUserIDErr(trans.ToID)
    }
  }

  switch trans.Type {
  case log_pb.Transaction_PUT:
  case log_pb.Transaction_DEPOSIT:
  case log_pb.Transaction_WITHDRAW:
    if ac.data[trans.UserID] < trans.Value {
      return secure.NewBalanceNotEnoughErr(trans.UserID, ac.data[trans.UserID],
                                           trans.Value)
    }
  case log_pb.Transaction_TRANSFER:
    if ac.data[trans.FromID] < trans.Value {
      return secure.NewBalanceNotEnoughErr(trans.FromID, ac.data[trans.FromID],
                                           trans.Value)
    }
  default:
    return secure.NewInvalidTypeErr()
  }

  return nil
}

func (ac *AccountStates) Print() {
  log.Println(ac.data)
}
