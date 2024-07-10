Compiler for hydrogen written in Python

Inspiration: https://github.com/orosmatthew/hydrogen-cpp

## Requirements

This compiler is for Linux, using nasm and the GNU linker.

## Instructions

1. Create a hydrogen file(.hy)

2. Call ```go run src/main.go <filename>.hy``` This will turn the hydrogen file into a .asm file.

3. Call ```nasm -felf64 <filename>.asm && ld <filename>.o -o <filename>``` This will create an object file, and then use the object file to create an executable.
4. Call ```./<filename>``` to run the executable.
