package sha3

import (
	"fmt"
	"testing"

	"github.com/ledgerwatch/erigon/crypto"
)

// This is similar to func rlpHash(x interface{}) (h libcommon.Hash) from core/types/hashing.go
func keccak_non_avx(data []byte) [32]byte {
	var h [32]byte
	sha := crypto.NewKeccakState()
	sha.Write(data)
	sha.Read(h[:])
	return h
}

func run_keccak(data [4][]byte) error {
	hashes, err := Hash_Keccak_AVX2(data)
	if err != nil {
		return err
	}
	fmt.Printf("AVX2 Hashes:\n%x\n%x\n%x\n%x\n", hashes[0], hashes[1], hashes[2], hashes[3])

	var expected [4][32]byte
	for i := 0; i < 4; i++ {
		expected[i] = keccak_non_avx(data[i])
	}
	fmt.Printf("Non-AVX Hashes:\n%x\n%x\n%x\n%x\n", expected[0], expected[1], expected[2], expected[3])

	for i := 0; i < 4; i++ {
		if expected[i] != hashes[i] {
			return fmt.Errorf("Expected: %x, got: %x", expected[i], hashes[i])
		}
	}

	return nil
}

func TestAVXKeccak(t *testing.T) {
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

	msgs3 := [4][]byte{
		[]byte("a short message"),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
		[]byte("7877910f6b0e828ddd54cb252919af164ab996a8e702de2181e0d2a86bdd2e8c"),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
	}

	err := run_keccak(msgs1)
	if err != nil {
		t.Fatal(err)
	}
	run_keccak(msgs2)
	if err != nil {
		t.Fatal(err)
	}
	run_keccak(msgs3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done.")
}
