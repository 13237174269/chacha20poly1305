package chacha20poly1305

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/codahale/chacha20"
	"github.com/codahale/poly1305"
	"testing"
)

// stolen from http://tools.ietf.org/html/draft-agl-tls-chacha20poly1305-02#section-7
var testVectors = [][]string{
	[]string{
		"4290bcb154173531f314af57f3be3b5006da371ece272afa1b5dbdd1100a1007",
		"86d09974840bded2a5ca",
		"cd7cf67be39c794a",
		"87e229d4500845a079c0",
		"e3e446f7ede9a19b62a4677dabf4e3d24b876bb284753896e1d6",
	},
}

func TestSealing(t *testing.T) {
	for i, vector := range testVectors {
		t.Logf("Running test vector %d", i)

		key, err := hex.DecodeString(vector[0])
		if err != nil {
			t.Error(err)
		}

		plaintext, err := hex.DecodeString(vector[1])
		if err != nil {
			t.Error(err)
		}

		nonce, err := hex.DecodeString(vector[2])
		if err != nil {
			t.Error(err)
		}

		data, err := hex.DecodeString(vector[3])
		if err != nil {
			t.Error(err)
		}

		expected, err := hex.DecodeString(vector[4])
		if err != nil {
			t.Error(err)
		}

		c, err := NewChaCha20Poly1305(key)
		if err != nil {
			t.Error(err)
		}

		actual := c.Seal(nil, nonce, plaintext, data)

		if !bytes.Equal(expected, actual) {
			t.Errorf("Bad seal: expected %x, was %x", expected, actual)

			for i, v := range expected {
				if actual[i] != v {
					t.Logf("Mismatch at offset %d: %x vs %x", i, v, actual[i])
					break
				}
			}
		}
	}
}

func TestRoundtrip(t *testing.T) {
	key := make([]byte, KeySize)

	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Error(err)
	}

	nonce := make([]byte, c.NonceSize())
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	ciphertext := c.Seal(nil, nonce, plaintext, data)

	actual, err := c.Open(nil, nonce, ciphertext, data)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plaintext, actual) {
		t.Errorf("Bad seal: expected %x, was %x", plaintext, actual)
	}
}

func TestModifiedData(t *testing.T) {
	key := make([]byte, KeySize)

	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Error(err)
	}

	nonce := make([]byte, c.NonceSize())
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	ciphertext := c.Seal(nil, nonce, plaintext, data)

	data[0] ^= 1

	_, err = c.Open(nil, nonce, ciphertext, data)
	if err != ErrAuthFailed {
		t.Error("Should have failed, but didn't")
	}
}

func TestModifiedCiphertext(t *testing.T) {
	key := make([]byte, KeySize)

	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Error(err)
	}

	nonce := make([]byte, c.NonceSize())
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	ciphertext := c.Seal(nil, nonce, plaintext, data)

	ciphertext[0] ^= 1

	_, err = c.Open(nil, nonce, ciphertext, data)
	if err != ErrAuthFailed {
		t.Error("Should have failed, but didn't")
	}
}

func TestNonceSize(t *testing.T) {
	key := make([]byte, KeySize)
	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Error(err)
	}

	if c.NonceSize() != chacha20.NonceSize {
		t.Errorf("Expected nonce size of %d but was %d", chacha20.NonceSize, c.NonceSize())
	}
}

func TestOverhead(t *testing.T) {
	key := make([]byte, KeySize)
	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Error(err)
	}

	if c.Overhead() != poly1305.Size {
		t.Errorf("Expected overhead of %d but was %d", poly1305.Size, c.Overhead())
	}
}

func TestInvalidKey(t *testing.T) {
	key := make([]byte, 31)
	_, err := NewChaCha20Poly1305(key)

	if err != ErrInvalidKey {
		t.Errorf("Expected invalid key error but was %v", err)
	}
}

func TestSealInvalidNonce(t *testing.T) {
	defer func() {
		if r := recover(); r != nil && r != ErrInvalidNonce {
			t.Errorf("Expected invalid key panic but was %v", r)
		}
	}()

	key := make([]byte, KeySize)
	c, err := NewChaCha20Poly1305(key)

	if err != nil {
		t.Error(err)
	}

	nonce := make([]byte, c.NonceSize()-3)
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	c.Seal(nil, nonce, plaintext, data)
}

func TestOpenInvalidNonce(t *testing.T) {
	key := make([]byte, KeySize)
	c, err := NewChaCha20Poly1305(key)

	if err != nil {
		t.Error(err)
	}

	nonce := make([]byte, c.NonceSize())
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	ciphertext := c.Seal(nil, nonce, plaintext, data)

	_, err = c.Open(nil, nonce[0:4], ciphertext, data)

	if err != ErrInvalidNonce {
		t.Errorf("Expected invalid nonce error but was %v", err)
	}
}

func BenchmarkChaCha20Poly1305(b *testing.B) {
	key := make([]byte, KeySize)
	nonce := make([]byte, 8)
	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		panic(err)
	}

	input := make([]byte, 1024*1024)

	b.SetBytes(int64(len(input)))
	for i := 0; i < b.N; i++ {
		c.Seal(nil, nonce, input, nil)
	}
}

func readSecretKey(i int) []byte {
	return make([]byte, i)
}

func readRandomNonce(i int) []byte {
	return make([]byte, i)
}

func ExampleNewChaCha20Poly1305() {
	key := readSecretKey(KeySize) // must be 256 bits long

	c, err := NewChaCha20Poly1305(key)
	if err != nil {
		panic(err)
	}

	nonce := readRandomNonce(c.NonceSize()) // must be generated by crypto/rand
	plaintext := []byte("yay for me")
	data := []byte("whoah yeah")
	ciphertext := c.Seal(nil, nonce, plaintext, data)

	fmt.Printf("%x\n", ciphertext)
	// Output:
	// e6669e9e333e4a5af5df2b8d1669cbdc175bb32da46484e6e358
}
