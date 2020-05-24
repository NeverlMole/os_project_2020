package secure

import (
  "fmt"
)

type InvalidUserIDErr struct {
  userID string
}

func NewInvalidUserIDErr(userID string) error {
  return &InvalidUserIDErr{userID}
}

func (e *InvalidUserIDErr) Error() string {
  return "The user ID [" + e.userID + "] is invalid."
}


type TransactionErr struct {
  transID int32
  reason error
}

func NewTransactionErr(transID int32, reason error) error {
  return &TransactionErr{transID, reason}
}

func (e *TransactionErr) Error() string {
  return fmt.Sprintf("Transaction %d failed with error: %v",
                     e.transID, e.reason)
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


type InvalidTypeErr struct {}

func NewInvalidTypeErr() error {
  return &InvalidTypeErr{}
}

func (e *InvalidTypeErr) Error() string {
  return "The type of transaction is invalid."
}

type UserGetErr struct {
  reason error
}


func NewUserGetErr(reason error) error {
  return &UserGetErr{reason}
}

func (e *UserGetErr) Error() string {
  return fmt.Sprintf("Get failed with error: %v", e.reason)
}

type WriteLogFailErr struct {
  id int32
  reason error
}


func NewWriteLogFailErr(id int32, reason error) error {
  return &WriteLogFailErr{id, reason}
}

func (e *WriteLogFailErr) Error() string {
  return fmt.Sprintf("Write log %d with error: %v", e.id, e.reason)
}

type WriteBlockFailErr struct {
  id int32
  reason error
}


func NewWriteBlockFailErr(id int32, reason error) error {
  return &WriteBlockFailErr{id, reason}
}

func (e *WriteBlockFailErr) Error() string {
  return fmt.Sprintf("Write block %d with error: %v", e.id, e.reason)
}

type BlockInvalidErr struct {
  id int32
}


func NewBlockInvalidErr(id int32) error {
  return &BlockInvalidErr{id}
}

func (e *BlockInvalidErr) Error() string {
  return fmt.Sprintf("The block %d is invalid.", e.id)
}


type ApplyInconsistentErr struct {}

func NewApplyInconsistentErr() error {
  return &ApplyInconsistentErr{}
}

func (e *ApplyInconsistentErr) Error() string {
  return "APPLY INCONSISTENT !!! The same transaction was able to apply " +
         "once but failed now."
}


type DataMissingErr struct {
  data string
}

func NewDataMissingErr(data string) error {
  return &DataMissingErr{data}
}

func (e *DataMissingErr) Error() string {
  return fmt.Sprintf("Some %s are missing.", e.data)
}


type LogsInvalidErr struct {
  id int32
  reason error
}

func NewLogsInvalidErr(id int32, reason error) error {
  return &LogsInvalidErr{id, reason}
}

func (e *LogsInvalidErr) Error() string {
  return fmt.Sprintf("The log %d is invalid with error: %v", e.id, e.reason)
}


type ServerInitErr struct {
  reason error
}

func NewServerInitErr(reason error) error {
  return &ServerInitErr{reason}
}

func (e *ServerInitErr) Error() string {
  return fmt.Sprintf("The server's recovery failed with error: %v", e.reason)
}
