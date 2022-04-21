//go:build !purego && (aix || android || darwin || dragonfly || freebsd || illumos || ios || linux || netbsd || openbsd || plan9 || solaris)
// +build !purego
// +build aix android darwin dragonfly freebsd illumos ios linux netbsd openbsd plan9 solaris

// TODO: replace the above with go:build unix once Go 1.19 is the lowest
// supported version

package buffer

import (
	"golang.org/x/sys/unix"
)

type Buffer struct {
	n    int
	pg   int
	mmap []byte
}

func New(n int) (Buffer, error) {
	pg := unix.Getpagesize()
	full := ((n+(pg-1))/pg + 2) * pg

	b, err := unix.Mmap(-1, 0, full, unix.PROT_NONE, unix.MAP_ANON|unix.MAP_PRIVATE)
	if err != nil {
		return Buffer{}, err
	}

	if n > 0 {
		err = unix.Mprotect(b[pg:full-pg], unix.PROT_READ|unix.PROT_WRITE)
		if err != nil {
			unix.Munmap(b)
			return Buffer{}, err
		}
	}

	return Buffer{
		n:    n,
		pg:   pg,
		mmap: b,
	}, nil
}

func (a *Buffer) ProtectHead() []byte {
	head := a.pg
	return a.mmap[head : head+a.n : head+a.n]
}

func (a *Buffer) ProtectTail() []byte {
	tail := len(a.mmap) - a.pg - a.n
	return a.mmap[tail : tail+a.n : tail+a.n]
}

func (a *Buffer) Release() {
	unix.Munmap(a.mmap)
}
