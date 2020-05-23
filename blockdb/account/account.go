package account

import (
  log_pb "blockdb/log_protobuf/go"
)

type AccountStates struct {
  data map[string]int32
}

func New() (*AccountStates) {
  return &AccountStates{make(map[string]int32)}
}

func (ac *AccountStates) Get(userID string) (int32, error) {
  /* TODO: check whether the userID is valid. If not return an error. */
  return ac.data[userID], nil
}

func (ac *AccountStates) Apply(trans log_pb.Transaction) (bool, error) {
  /* TODO: check whether the userID is valid. If not return an error. */
  switch trans.Type {
  case log_pb.Transaction_PUT:
    ac.data[trans.UserID] = trans.Value
  case log_pb.Transaction_DEPOSIT:
    ac.data[trans.UserID] += trans.Value
  case log_pb.Transaction_WITHDRAW:
    if ac.data[trans.UserID] < trans.Value {
      /* TODO: return BalanceNotEnoughErr. */
      return false, nil
    }
    ac.data[trans.UserID] -= trans.Value
  case log_pb.Transaction_TRANSFER:
    if ac.data[trans.FromID] < trans.Value {
      /* TODO: return BalanceNotEnoughErr. */
      return false, nil
    }
    ac.data[trans.FromID] -= trans.Value
    ac.data[trans.ToID] += trans.Value
  default:
    /* TODO: return InvalidTransactionErr. */
    return false, nil
  }

  return true, nil
}
