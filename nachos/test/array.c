#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define SIZE 10

int arr[SIZE]={0};

int main()
{
  printf(arr[8]);
  printf(arr[20]);
  printf(arr[20000]);
  
  return 0;
}
