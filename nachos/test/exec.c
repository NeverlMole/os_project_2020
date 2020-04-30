#include "syscall.h"

int main()
{
    char file1[100] = "halt.coff";
    int args = 0;
    char *argv[10];

    exec(file1, args, argv);


    char file2[100] = "echo.coff";
    args = 5;
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

    exec(file2, args, argv);
}
