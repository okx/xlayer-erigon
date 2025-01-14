package sha3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/simd/keccakf1600"
	"github.com/ledgerwatch/erigon/core/types"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
)

func populateState(data []byte, idx int, state *[]uint64, pstart *int) {
	rate := 136
	rate_64 := rate / 8

	start := *pstart
	if start+rate <= len(data) {
		for j := 0; j < rate_64; j++ {
			(*state)[4*j+idx] = binary.LittleEndian.Uint64(data[start+j*8 : start+(j+1)*8])
		}
		*pstart = start + rate
	} else {
		rem_64 := (len(data) - start) / 8
		rem_bytes := (len(data) - start) % 8
		for j := 0; j < rem_64; j++ {
			(*state)[4*j+idx] = binary.LittleEndian.Uint64(data[start+j*8 : start+(j+1)*8])
		}
		if rem_bytes > 0 {
			buf := make([]byte, 8)
			copy(buf[:], data[start+rem_64*8:])
			(*state)[4*rem_64+idx] = binary.LittleEndian.Uint64(buf) | (uint64(0x1) << (rem_bytes * 8))
		} else {
			(*state)[4*rem_64+idx] = 0x1
		}
		(*state)[(rate_64-1)*4+idx] = 0x80 << 56
		*pstart = len(data)
	}
}

func populateHash(state []uint64, idx int, hashes *[32]byte) {
	for j := 0; j < 4; j++ {
		binary.LittleEndian.PutUint64(
			(*hashes)[8*j:8*(j+1)],
			state[4*j+idx],
		)
	}
}

func Hash_Keccak_AVX2(data [4][]byte) ([4][32]byte, error) {
	if !keccakf1600.IsEnabledX4() {
		return [4][32]byte{}, errors.New("AVX2 not available")
	}

	var hashes [4][32]byte

	// f1600 acts on 1600 bits arranged as 25 uint64s.  Our fourway f1600
	// acts on four interleaved states; that is a [100]uint64.  (A separate
	// type is used to ensure that the encapsulated [100]uint64 is aligned
	// properly to be used efficiently with vector instructions.)
	var perm keccakf1600.StateX4
	state := perm.Initialize(false)

	// state is initialized with zeroes
	start1 := 0
	start2 := 0
	start3 := 0
	start4 := 0
	done1 := false
	done2 := false
	done3 := false
	done4 := false

	for !done1 || !done2 || !done3 || !done4 {
		populateState(data[0], 0, &state, &start1)
		populateState(data[1], 1, &state, &start2)
		populateState(data[2], 2, &state, &start3)
		populateState(data[3], 3, &state, &start4)
		perm.Permute()
		if start1 == len(data[0]) && !done1 {
			done1 = true
			populateHash(state, 0, &hashes[0])
		}
		if start2 == len(data[1]) && !done2 {
			done2 = true
			populateHash(state, 1, &hashes[1])
		}
		if start3 == len(data[2]) && !done3 {
			done3 = true
			populateHash(state, 2, &hashes[2])
		}
		if start4 == len(data[3]) && !done4 {
			done4 = true
			populateHash(state, 3, &hashes[3])
		}
	}

	return hashes, nil
}

func RlpHashAVX2(x []*types.Transaction) []libcommon.Hash {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer
	var buf3 bytes.Buffer
	var buf4 bytes.Buffer
	if x[0] == nil {
		buf1.Write([]byte{0x0})
	} else {
		(*x[0]).EncodeRLP(&buf1)
	}
	if x[1] == nil {
		buf2.Write([]byte{0x0})
	} else {
		(*x[1]).EncodeRLP(&buf2)
	}
	if x[2] == nil {
		buf3.Write([]byte{0x0})
	} else {
		(*x[2]).EncodeRLP(&buf3)
	}
	if x[3] == nil {
		buf4.Write([]byte{0x0})
	} else {
		(*x[3]).EncodeRLP(&buf4)
	}
	hashes, err := Hash_Keccak_AVX2([4][]byte{buf1.Bytes(), buf2.Bytes(), buf3.Bytes(), buf4.Bytes()})
	if err != nil {
		fmt.Errorf("RlpHashAVX2() Error: %v", err)
		return nil
	}
	h := make([]libcommon.Hash, 4)
	h[0] = libcommon.BytesToHash(hashes[0][:])
	h[1] = libcommon.BytesToHash(hashes[1][:])
	h[2] = libcommon.BytesToHash(hashes[2][:])
	h[3] = libcommon.BytesToHash(hashes[3][:])
	return h
}
