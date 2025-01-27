//go:build amd64 && !purego
// +build amd64,!purego

package keccak

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

// ! Taken from https://github.com/cloudflare/circl/blob/main/simd/keccakf1600/f1600x.go

// StateX4 contains state for the four-way permutation including the four
// interleaved [25]uint64 buffers. Call Initialize() before use to initialize
// and get a pointer to the interleaved buffer.
type StateX4 struct {
	// Go guarantees a to be aligned on 8 bytes, whereas we need it to be
	// aligned on 32 bytes for bet performance.  Thus we leave some headroom
	// to be able to move the start of the state.

	// 4 x 25 uint64s for the interleaved states and three uint64s headroom
	// to fix alignment.
	a [103]uint64

	// Offset into a that is 32 byte aligned.
	offset int

	// If true, permute will use 12-round keccak instead of 24-round keccak
	turbo bool
}

// IsEnabledX4 returns true if the architecture supports a four-way SIMD
// implementation provided in this package.
func IsEnabledX4() bool { return cpu.X86.HasAVX2 }

// Initialize the state and returns the buffer on which the four permutations
// will act: a uint64 slice of length 100.  The first permutation will act
// on {a[0], a[4], ..., a[96]}, the second on {a[1], a[5], ..., a[97]}, etc.
// If turbo is true, applies 12-round variant instead of the usual 24.
func (s *StateX4) Initialize(turbo bool) []uint64 {
	s.turbo = turbo
	rp := unsafe.Pointer(&s.a[0])

	// uint64s are always aligned by a multiple of 8.  Compute the remainder
	// of the address modulo 32 divided by 8.
	rem := (int(uintptr(rp)&31) >> 3)

	if rem != 0 {
		s.offset = 4 - rem
	}

	// The slice we return will be aligned on 32 byte boundary.
	return s.a[s.offset : s.offset+100]
}

// Permute performs the four parallel Keccak-f[1600]s interleaved on the slice
// returned from Initialize().
func (s *StateX4) Permute() {
	permuteSIMDx4(s.a[s.offset:], s.turbo)
}

func permuteSIMDx4(state []uint64, turbo bool) { f1600x4AVX2(&state[0], &RC, turbo) }

func f1600x4AVX2(state *uint64, rc *[24]uint64, turbo bool)
