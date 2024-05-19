#include "stdio.h"

int x;

int add(int a, int b) {
    return a + b;
}

int main() {
    printf("Hello, world!\n");
    x = add(5, 3);
    printf("5 + 3 = %d\n", x);
    return 0;
}
