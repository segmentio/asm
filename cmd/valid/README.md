# valid

This program is a helper to check the output of `utf8.Valid` facilitate
debugging. It accepts some input, runs both this library and stdlib's version of
`utf8.Valid`, and prints out the result.

## Usage

Provide the input as the the first argument to the program:

```
$ go run main.go "hello! 😊"
hello! 😊
[104 101 108 108 111 33 32 240 159 152 138]
11 bytes
stdlib: utf8: true ascii: false
valid:  utf8: true ascii: false v: 1
```

The input is parsed as a double quoted Go string, so you can use escape codes:

```
$ go run main.go "\xFA"

[250]
1 bytes
stdlib: utf8: false ascii: false
valid:  utf8: false ascii: false v: 0
```

Alternatively it can also conusme input from stdin:

```
$ cat example.txt
hello! 😊
$ go run main.go < example.txt
hello! 😊

[104 101 108 108 111 33 32 240 159 152 138 10]
12 bytes
stdlib: utf8: true ascii: false
valid:  utf8: true ascii: false v: 1
```

As a bonus, if the file is the result of a failure reported by Go 1.18 fuzz, the
program extracts the actual value of the test:

```
$ cat fuzz.out
go test fuzz
[]byte("000000000000000000~\xFF")
$ go run main.go < fuzz.out
Got fuzzer input
000000000000000000~
[48 48 48 48 48 48 48 48 48 48 48 48 48 48 48 48 48 48 126 255]
20 bytes
stdlib: utf8: false ascii: false
valid:  utf8: false ascii: false v: 0
```

## GDB

A useful way to debug is to run this program with some problematic input and use
GDB to step through the execution and inspect registers. The `debug.gdb` file is
a basic helper to automate part of the process. For example:

```
$ go build main.go && gdb --command=debug.gdb  -ex "set args < ./example.txt" ./main
```
