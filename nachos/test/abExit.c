#include "syscall.h"

int main(int argc, char** argv)
{
  char bite[10];
  read(0, bite, 1);
  exit(argc);
    /* not reached */
}
