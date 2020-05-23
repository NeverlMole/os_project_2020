package util

import (
  "os"
)

/* Check whether a directory or path exists. */
func PathIsExist(dir string) bool {
  if _, err := os.Stat(dir); os.IsNotExist(err) {
    return false
  }
  return true
}

/* To a path into the form end by "/" */
func Directorize(dir string) string {
  if dir[len(dir) - 1:] != "/" {
    dir = dir + "/"
  }
  return dir
}
