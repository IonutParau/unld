#!/usr/bin/env python3
import os

exe = "unld"
cc = "gcc"

print("Testing basic extraction")
if os.system(f"{cc} -o test testfiles/test.c"):
    print("Failed to generate test executable")
    exit(1)
if os.system(f"./{exe} test --empty -a add -o libadd.o"):
    os.remove("test")
    print("Failed to unlink executable")
    exit(1)
if os.system(f"{cc} -o rebuilt libadd.o testfiles/reconstructed.c"):
    os.remove("libadd.o")
    os.remove("test")
    print("Failed to rebuild executable with extracted library")
    exit(1)
if os.system(f"./rebuilt"):
    os.remove("libadd.o")
    os.remove("test")
    os.remove("rebuilt")
    print("Rebuilt binary does not work")
    exit(1)

os.remove("libadd.o")
os.remove("test")
os.remove("rebuilt")
print("Basic extration works")
