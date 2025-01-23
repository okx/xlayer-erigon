package sha3

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

// StateX8 contains state for the eight-way permutation including the eight
// interleaved [25]uint64 buffers. Call Initialize() before use to initialize
// and get a pointer to the interleaved buffer.
type StateX8 struct {
	// Go guarantees a to be aligned on 8 bytes, whereas we need it to be
	// aligned on 64 bytes for bet performance.  Thus we leave some headroom
	// to be able to move the start of the state.

	// 8 x 25 uint64s for the interleaved states and seven uint64s headroom
	// to fix alignment.
	a [207]uint64

	// Offset into a that is 64 byte aligned.
	offset int

	// If true, permute will use 12-round keccak instead of 24-round keccak
	turbo bool
}

func IsEnabledX8() bool { return cpu.X86.HasAVX512 }

// Initialize the state and returns the buffer on which the eight permutations
// will act: a uint64 slice of length 200. The first permutation will act
// on {a[0], a[8], ..., a[192]}, the second on {a[1], a[9], ..., a[193]}, etc.
// If turbo is true, applies 12-round variant instead of the usual 24.
func (s *StateX8) Initialize(turbo bool) []uint64 {
	s.turbo = turbo
	rp := unsafe.Pointer(&s.a[0])

	// Check alignment for 64 bytes and calculate the remainder
	rem := (int(uintptr(rp)&63) >> 3) // 64-byte alignment, unit: uint64 (8 bytes)

	if rem != 0 {
		// Adjust the offset to align on a 64-byte boundary
		s.offset = 8 - rem // Offset is the number of uint64s (8 bytes) needed
	}

	// The slice we return will be aligned on a 64-byte boundary
	return s.a[s.offset : s.offset+200]
}

var RC = [24]uint64{
	0x0000000000000001,
	0x0000000000008082,
	0x800000000000808A,
	0x8000000080008000,
	0x000000000000808B,
	0x0000000080000001,
	0x8000000080008081,
	0x8000000000008009,
	0x000000000000008A,
	0x0000000000000088,
	0x0000000080008009,
	0x000000008000000A,
	0x000000008000808B,
	0x800000000000008B,
	0x8000000000008089,
	0x8000000000008003,
	0x8000000000008002,
	0x8000000000000080,
	0x000000000000800A,
	0x800000008000000A,
	0x8000000080008081,
	0x8000000000008080,
	0x0000000080000001,
	0x8000000080008008,
}

func (s *StateX8) Permute() {
	permuteSIMDx8(s.a[s.offset:], s.turbo)
}

func permuteSIMDx8(state []uint64, turbo bool) { f1600x8AVX512(&state[0], &RC, turbo) }

func f1600x8AVX512(state *uint64, rc *[24]uint64, turbo bool)
