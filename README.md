# asm ![build status](https://github.com/segmentio/asm/actions/workflows/go.yml/badge.svg) [![GoDoc](https://godoc.org/github.com/segmentio/asm?status.svg)](https://godoc.org/github.com/segmentio/asm)

Go library providing algorithms optimized to leverage the characteristics of
modern CPUs.

## Motivation

With the development of Cloud technologies, access to large scale compute
capacity has never been easier, and running distributed systems deployed across
dozens or sometimes hundreds of CPUs has become common practice. As a side
effect of being provided seemingly unlimited (but somewhat expensive) compute
capacity, software engineers are now in direct connections with the economical
and environmental impact of running the software they develop in production;
performance and efficiency of our programs matters today more than it has ever
before.

Modern CPUs are complex machines with performance characteristic that may
vary by orders of magnitude depending on how they are used. Features like
branch prediction, instruction reordering, pipelining, or caching are all
input variables that determine the compute throughput that a CPU can achieve.
While compilers keep being improved, and often employ micro-optimizations that
would be counter-productive for human developers to be responsible for, there
are limitations to what they can do, and Assembly still has a role to play in
optimizing algorithms on hot code paths of large scale applications.

SIMD instruction sets offer interesting opportunities for software engineers.
Taking advantage of these instructions often requires rethinking how the program
represents and manipulates data, which is beyond the realm of optimizations that
can be implemented by a compiler. When renting CPU time from a Cloud provider,
programs that fail to leverage the full sets of instructions available are
therefore paying for features they do not use.

This package aims to provide such algorithms, optimized to leverage advanced
instruction sets of modern CPUs to maximize throughput and take the best
advantage of the available compute power. Users of the package will find
functions that have often been designed to work on **arrays of values**,
which is where SIMD and branchless algorithms shine.

The functions in this library have been used in high throughput production
environments at Segment, we hope that they will be useful to other developers
using Go in performance-sensitive software.

## Usage

The library is composed of multiple Go packages intended to act as logical
groups of functions sharing similar properties:

| Package | Purpose |
| ------- | ------- |
| `github.com/segmentio/asm/ascii` | library of functions designed to work on ASCII inputs |
| `github.com/segmentio/asm/base64` | standard library compatible base64 encodings |
| `github.com/segmentio/asm/bswap` | byte swapping algorithms working on arrays of fixed-size items |
| `github.com/segmentio/asm/cpu` | definition of the ABI used to detect CPU features |
| `github.com/segmentio/asm/mem` | functions operating on byte arrays |
| `github.com/segmentio/asm/qsort` | quick-sort implementations for arrays of fixed-size items |
| `github.com/segmentio/asm/slices` | functions performing computations on pairs of slices |
| `github.com/segmentio/asm/sortedset` | functions working on sorted arrays of fixed-size items |

When no assembly version of a function is available for the target platform,
the package provides a generic implementation in Go which is automatically
picked up by the compiler.

## Maintenance

Generation of the assembly code is managed with [AVO](https://github.com/mmcloughlin/avo),
and orchestrated by a Makefile which helps maintainers rebuild the assembly
source code when the AVO files are modified.

The repository contains two Go modules; the main module is declared as
`github.com/segmentio/asm` at the root of the repository, and the second
module is found in the `build` subdirectory.

The `build` module is used to isolate build dependencies from programs that
import the main module. Through this mechanism, AVO does not become a
dependency of programs using `github.com/segmentio/asm`, keeping the
dependency management overhead minimal for the users, and allowing
maintainers to make modifications to the `build` package.

Versioning of the two modules is managed independently; while we aim to provide
stable APIs on the main package, breaking changes may be introduced on the
`build` package more often, as it is intended to be ground for more experimental
constructs in the project.
