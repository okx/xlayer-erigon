package sha3

import (
	"fmt"
	"sync"
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

func run_bench_non_avx() {
	msg := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	var wg sync.WaitGroup
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			keccak_non_avx(msg)
		}(i)
	}
	wg.Wait()
}

func run_bench_avx() {
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
			Hash_Keccak_AVX2(data)
		}(i)
	}
	wg.Wait()
}
func BenchmarkNonAVXKeccak(b *testing.B) {
	for n := 0; n < b.N; n++ {
		run_bench_non_avx()
	}
}

func BenchmarkAVXKeccak(b *testing.B) {
	for n := 0; n < b.N; n++ {
		run_bench_avx()
	}
}

func run_keccak_avx_512(data [8][]byte) error {
	hashes, err := Hash_Keccak_AVX512(data)
	if err != nil {
		return err
	}
	fmt.Printf("AVX512 Hashes:\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n", hashes[0], hashes[1], hashes[2], hashes[3], hashes[4], hashes[5], hashes[6], hashes[7])

	var expected [8][32]byte
	for i := 0; i < 8; i++ {
		expected[i] = keccak_non_avx(data[i])
	}
	fmt.Printf("Non-AVX Hashes:\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n", expected[0], expected[1], expected[2], expected[3], expected[4], expected[5], expected[6], expected[7])
	for i := 0; i < 4; i++ {
		if expected[i] != hashes[i] {
			return fmt.Errorf("Expected: %x, got: %x", expected[i], hashes[i])
		}
	}
	return nil
}

func TestAVXKeccak_512(t *testing.T) {
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
		[]byte("a short message"),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
		[]byte("7877910f6b0e828ddd54cb252919af164ab996a8e702de2181e0d2a86bdd2e8c"),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
		[]byte("a short message"),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
		[]byte("7877910f6b0e828ddd54cb252919af164ab996a8e702de2181e0d2a86bdd2e8c"),
		[]byte("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
	}
	err := run_keccak_avx_512(msgs1)
	if err != nil {
		t.Fatal(err)
	}
	err = run_keccak_avx_512(msgs2)
	if err != nil {
		t.Fatal(err)
	}
	err = run_keccak_avx_512(msgs3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done.")

}
