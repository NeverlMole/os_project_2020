package secure

import (
  "fmt"
)


type BlockMapBrokenErr struct {}

func NewBlockMapBrokenErr() error {
  return &BlockMapBrokenErr{}
}

func (e *BlockMapBrokenErr) Error() string {
  return "The block map is broken."
}


type InvalidUserIDErr struct {
  userID string
}

func NewInvalidUserIDErr(userID string) error {
  return &InvalidUserIDErr{userID}
}

func (e *InvalidUserIDErr) Error() string {
  return "The user ID [" + e.userID + "] is invalid."
}


type InvalidServerIDErr struct {
  serverID string
}

func NewInvalidServerIDErr(serverID string) error {
  return &InvalidServerIDErr{serverID}
}

func (e *InvalidServerIDErr) Error() string {
  return "The server ID [" + e.serverID + "] is invalid."
}


type TransactionErr struct {
  UUID string
  reason error
}

func NewTransactionErr(UUID string, reason error) error {
  return &TransactionErr{UUID, reason}
}

func (e *TransactionErr) Error() string {
  return fmt.Sprintf("Transaction [%s] failed with error: %v",
                     e.UUID, e.reason)
}


type BalanceNotEnoughErr struct {
  userID string
  balance int32
  amount int32
}

func NewBalanceNotEnoughErr(userID string, balance int32, amount int32) error {
  return &BalanceNotEnoughErr{userID, balance, amount}
}

func (e *BalanceNotEnoughErr) Error() string {
  return fmt.Sprintf(
    "The user [%s]'s balance (%d) is not enough for amount (%d).",
    e.userID, e.balance, e.amount)
}


type IntegrityConstrainErr struct {
  miningFee int32
  value int32
}

func NewIntegrityConstrainErr(miningFee int32, value int32) error {
  return &IntegrityConstrainErr{miningFee, value}
}

func (e *IntegrityConstrainErr) Error() string {
  return fmt.Sprintf(
    "The integrity constrain is not satisfied since mining fee is %d and value is %d.",
    e.miningFee, e.value)
}


type InvalidBlockErr struct {
  blockHash string
  reason error
}

func NewInvalidBlockErr(blockHash string, reason error) error {
  return &InvalidBlockErr{blockHash, reason}
}

func (e *InvalidBlockErr) Error() string {
  return fmt.Sprintf("The block with hash [%s] is invalid with error: %v",
    e.blockHash, e.reason)
}


type AddBlockErr struct {
  reason string
}

func NewAddBlockErr(reason string) error {
  return &AddBlockErr{reason}
}

func (e *AddBlockErr) Error() string {
  return fmt.Sprintf("Cannot add the block since %s", e.reason)
}
