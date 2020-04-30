#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE]="Hello World!\0";

int main()
{
  int file, amount=10;

  file = creat("a.txt");
  if (file==-1) {
    printf("Unable to open");
    return 1;
  }
 
 
  write(file, buf, strlen(buf));
  close(file);
  file = open("a.txt");

  while ((amount = read(file, buf, amount))>0) {
    write(1, buf, amount);
  }

  return 0;
}
