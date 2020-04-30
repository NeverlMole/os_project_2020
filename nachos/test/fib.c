#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 20

int fib[BUFSIZE], i;

int main()
{
  fib[0] = 1;
  fib[1] = 1;
  printf("Fib: 0, 1, ");
  for(i=2; i<BUFSIZE; i++){
    fib[i] = fib[i-1] + fib[i-2];
    printf("%d, ", fib[i]);
  }
  printf("and so on.");
  return 0;
}
