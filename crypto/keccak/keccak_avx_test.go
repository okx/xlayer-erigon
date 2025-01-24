package keccak

import (
	"fmt"
	"sync"
	"testing"

	"github.com/holiman/uint256"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/crypto"
)

// This is similar to func rlpHash(x interface{}) (h libcommon.Hash) from core/types/hashing.go
func keccakNonAVX(data []byte) [32]byte {
	var h [32]byte
	sha := crypto.NewKeccakState()
	sha.Write(data)
	sha.Read(h[:])
	return h
}

func runKeccakAVX2(data [4][]byte) error {
	hashes, err := HashKeccakAVX2(data)
	if err != nil {
		return err
	}
	fmt.Printf("AVX2 Hashes:\n%x\n%x\n%x\n%x\n", hashes[0], hashes[1], hashes[2], hashes[3])

	var expected [4][32]byte
	for i := 0; i < 4; i++ {
		expected[i] = keccakNonAVX(data[i])
	}
	fmt.Printf("Non-AVX Hashes:\n%x\n%x\n%x\n%x\n", expected[0], expected[1], expected[2], expected[3])

	for i := 0; i < 4; i++ {
		if expected[i] != hashes[i] {
			return fmt.Errorf("Hash %v: Expected: %x, got: %x", i, expected[i], hashes[i])
		}
	}
	return nil
}

func runKeccakAVX512(data [8][]byte) error {
	hashes, err := HashKeccakAVX512(data)
	if err != nil {
		return err
	}
	fmt.Printf("AVX512 Hashes:\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n", hashes[0], hashes[1], hashes[2], hashes[3], hashes[4], hashes[5], hashes[6], hashes[7])

	var expected [8][32]byte
	for i := 0; i < 8; i++ {
		expected[i] = keccakNonAVX(data[i])
	}
	fmt.Printf("Non-AVX Hashes:\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n", expected[0], expected[1], expected[2], expected[3], expected[4], expected[5], expected[6], expected[7])
	for i := 0; i < 8; i++ {
		if expected[i] != hashes[i] {
			return fmt.Errorf("Hash %v: Expected: %x, got: %x", i, expected[i], hashes[i])
		}
	}
	return nil
}

func runBenchNonAVX() {
	msg := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	var wg sync.WaitGroup
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			keccakNonAVX(msg)
		}(i)
	}
	wg.Wait()
}

func runBenchAVX2() {
	msg := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	var data [4][]byte
	for i := 0; i < 4; i++ {
		data[i] = msg
	}
	var wg sync.WaitGroup
	for i := 0; i < 1024; i += 4 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			HashKeccakAVX2(data)
		}(i)
	}
	wg.Wait()
}

func runBenchAVX512() {
	msg := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	var data [8][]byte
	for i := 0; i < 8; i++ {
		data[i] = msg
	}
	var wg sync.WaitGroup
	for i := 0; i < 1024; i += 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			HashKeccakAVX512(data)
		}(i)
	}
	wg.Wait()
}

func BenchmarkKeccakNonAVX(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runBenchNonAVX()
	}
}

func BenchmarkKeccakAVX2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runBenchAVX2()
	}
}

func BenchmarkKeccakAVX512(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runBenchAVX512()
	}
}

func TestKeccakAVX2(t *testing.T) {
	// Taken from: https://github.com/cloudflare/circl/blob/main/simd/keccakf1600/example_test.go

	msgs1 := [4][]byte{
		[]byte("These are some short"),
		[]byte("strings of the same "),
		[]byte("length that fit in a"),
		[]byte("single block.       "),
	}

	msgs2 := [4][]byte{
		[]byte("These are some short"),
		[]byte("strings of the same"),
		[]byte("length that fit in a"),
		[]byte("single block."),
	}

	// some corner cases
	msgs3 := [4][]byte{
		// len is 135
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e       "),
		// len is 136
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e        "),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
	}

	err := runKeccakAVX2(msgs1)
	if err != nil {
		t.Fatal(err)
	}
	err = runKeccakAVX2(msgs2)
	if err != nil {
		t.Fatal(err)
	}
	err = runKeccakAVX2(msgs3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done.")
}

func TestKeccakAVX512(t *testing.T) {
	msgs1 := [8][]byte{
		[]byte("These are some short"),
		[]byte("strings of the same "),
		[]byte("length that fit in a"),
		[]byte("single block.       "),
		[]byte("These are some short"),
		[]byte("strings of the same "),
		[]byte("length that fit in a"),
		[]byte("single block.       "),
	}

	msgs2 := [8][]byte{
		[]byte("These are some short"),
		[]byte("strings of the same"),
		[]byte("length that fit in a"),
		[]byte("single block."),
		[]byte("These are some short"),
		[]byte("strings of the same"),
		[]byte("length that fit in a"),
		[]byte("single block."),
	}

	msgs3 := [8][]byte{
		// len < 136
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e"),
		// len == 136
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e        "),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad "),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur? Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
		[]byte(" "),
	}

	err := runKeccakAVX512(msgs1)
	if err != nil {
		t.Fatal(err)
	}
	err = runKeccakAVX512(msgs2)
	if err != nil {
		t.Fatal(err)
	}

	err = runKeccakAVX512(msgs3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done.")
}

/**
 * TestAVXTxHash tests the AVX2 implementation of Keccak hashing on transactions (LeggacyTx).
 */
func TestTxHashAVX(t *testing.T) {
	tx := &types.LegacyTx{
		CommonTx: types.CommonTx{
			ChainID: uint256.NewInt(1),
			Nonce:   1024,
			Gas:     1000000,
			To:      &libcommon.Address{0x1},
			Value:   uint256.NewInt(1000000000),
			Data:    []byte("Hello, World!"),
			V:       *uint256.NewInt(128), // make it protected
		},
		GasPrice: uint256.NewInt(1000000000),
	}

	expected := tx.Hash()

	txs := [4]types.Transaction{tx, tx, tx, tx}
	var hashes [4]libcommon.Hash

	// test RlpHashAVX2
	err := RlpHashAVX(txs[:], hashes[:], false)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		if expected != hashes[i] {
			t.Fatalf("RlpHashAVX2 Hash %d -> Expected: %x, got: %x", i, expected, hashes[i])
		}
	}

	// test SigningHashAVX2 for ChainID != nil
	if !tx.Protected() {
		t.Fatalf("Transaction is not protected")
	}
	cid := uint256.NewInt(3)
	expected = tx.SigningHash(cid.ToBig())
	err = SigningHashAVX(cid.ToBig(), txs[:], hashes[:], false)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if expected != hashes[i] {
			t.Fatalf("SigningHashAVX2 Hash %d -> Expected: %x, got: %x", i, expected, hashes[i])
		}
	}

	// test SigningHashAVX2 for ChainID = nil
	expected = tx.SigningHash(nil)
	err = SigningHashAVX(nil, txs[:], hashes[:], false)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if expected != hashes[i] {
			t.Fatalf("SigningHashAVX2 Hash %d -> Expected: %x, got: %x", i, expected, hashes[i])
		}
	}
	fmt.Println("Done.")
}
