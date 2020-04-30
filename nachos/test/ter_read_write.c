#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char c;

int main()
{
  c='0';

  while (c!='q') {
    c=getchar();
    printf("%c", c);
  }

  return 0;
}
