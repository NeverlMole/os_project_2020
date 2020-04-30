#include "syscall.h"

int main()
{
    char file1[100] = "exit.coff";
    char file2[100] = "abExit.coff";
    int args = 0;
    char *argv[10];

    char arg1[] = "obn5jk43b274";
    char arg2[] = "5 n3kl21n";
    char arg3[] = "gf0dsa0";
    char arg4[] = "n4o32n1o";
    char arg5[] = "410uefg";
    argv[0] = arg1;
    argv[1] = arg2;
    argv[2] = arg3;
    argv[3] = arg4;
    argv[4] = arg5;

    int c1 = exec(file1, 0, argv);
    int c2 = exec(file1, 1, argv);
    int c3 = exec(file1, 2, argv);
    int c4 = exec(file2, 0, argv);

    printf("c1:%d, c2:%d, c3:%d, c4:%d\n", c1, c2, c3, c4);

    int status, returnValue;
    returnValue = join(c1, &status);
    printf("join c1 with status: %d and return value: %d\n", status,
           returnValue);

    returnValue = join(c2, &status);
    printf("join c2 with status: %d and return value: %d\n", status,
           returnValue);

    returnValue = join(c3, &status);
    printf("join c3 with status: %d and return value: %d\n", status,
           returnValue);


    returnValue = join(c4, &status);
    printf("join c4 with status: %d and return value: %d\n", status,
           returnValue);

    returnValue = join(-1, &status);
    printf("join -1 with status: %d and return value: %d\n", status,
           returnValue);
}
