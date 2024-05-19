# unld

Split an executable into multiple object files.

# Usage

CLI flags are explained when calling it with no arguments.

Here's a brief example:
Suppose we compiled this code
```c
#include "stdio.h"

int add(int x, int y) {
    return x + y;
}

int main() {
    printf("5 + 3 = %d\n", add(5, 3));
    return 0;
}
```
into `my_app`.
We can use
```sh
unld my_app --empty -a add -o libadd.o
```
This will generate a file, `libadd.o`, which contains an add symbol with the code taken from `my_app`.
This will include all necessary symbols from `.rodata`, which contains string literals most often, automatically.
This will NOT include the definitions for the symbols from `.bss` automatically. It will define them as external globals by default.
The flag `-g` will include the definition of the global if this object file is supposed to define it.

## Note

When linking it back, it is important to know that sometimes, the object file is not position independent.
This may require you use the flag `-no-pie` with ld or gcc, otherwise you may get linker errors.

# How it works

First, it uses objdump to disassemble the executable and take out all linking information.
Then, it will construct a representation of the entire executable as an object file.
Then, the CLI flags you pass in will operate on this object file model to output an assembly file in your temporary directory.
Then, it will call `nasm` to create an elf64 object file at the path requested, generating the output object file.
