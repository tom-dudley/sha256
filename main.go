package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func main() {
	var message []byte
	if len(os.Args) != 2 {
		fmt.Println("Usage: main.go [path]")
		os.Exit(1)
	}

	message, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %s", err)
		os.Exit(1)
	}

	messageDigest := sha256Digest(message)
	fmt.Println(messageDigest)
}

func sha256Digest(message []byte) string {
	// Key steps:
	// - Preprocess
	//   - Pad the message
	//   - Parse the message
	// - Hash

	// Step 1: Pad the message
	paddedMessage := padMessage([]byte(message))

	// Step 2: Parse the message
	parsedMessage := parseMessage(paddedMessage)

	// Step 3: Perform hashing
	// See Section 6.2.2 of FIPS 180-4
	hashValueWords := initialHashValues()

	for i := 0; i < len(parsedMessage); i++ {
		// 1. Prepare the message schedule
		w := make([]uint32, 64, 64)
		for t := 0; t < 64; t++ {
			if t < 16 {
				// each message block is 512 bits long (64 bytes)
				// the message schedule has 64 entries, each being a 'word'
				// each word is 32 bits
				w[t] = getWordFromMessageBlock(parsedMessage[i], t)
			} else {
				w[t] = lowerSigma1(w[t-2]) + w[t-7] + lowerSigma0(w[t-15]) + w[t-16]
			}
		}

		// 2. Initialise the eight working variables
		a := hashValueWords[0]
		b := hashValueWords[1]
		c := hashValueWords[2]
		d := hashValueWords[3]
		e := hashValueWords[4]
		f := hashValueWords[5]
		g := hashValueWords[6]
		h := hashValueWords[7]

		// 3.
		for t := 0; t < 64; t++ {
			T1 := h + upperSigma1(e) + ch(e, f, g) + kConstant(t) + w[t]
			T2 := upperSigma0(a) + maj(a, b, c)
			h = g
			g = f
			f = e
			e = d + T1
			d = c
			c = b
			b = a
			a = T1 + T2
		}

		// 4. Compute the ith intermediate ash value
		hashValueWords[0] = a + hashValueWords[0]
		hashValueWords[1] = b + hashValueWords[1]
		hashValueWords[2] = c + hashValueWords[2]
		hashValueWords[3] = d + hashValueWords[3]
		hashValueWords[4] = e + hashValueWords[4]
		hashValueWords[5] = f + hashValueWords[5]
		hashValueWords[6] = g + hashValueWords[6]
		hashValueWords[7] = h + hashValueWords[7]
	}

	digest := fmt.Sprintf("%x%x%x%x%x%x%x%x",
		hashValueWords[0],
		hashValueWords[1],
		hashValueWords[2],
		hashValueWords[3],
		hashValueWords[4],
		hashValueWords[5],
		hashValueWords[6],
		hashValueWords[7])

	return digest
}

func getWordFromMessageBlock(messageBlock []byte, i int) uint32 {
	if len(messageBlock) != 64 {
		panic("Message block is not 512 bits in length")
	}

	if i >= 16 {
		panic("Cannot fetch word greater than index 15")
	}

	b0 := messageBlock[i*4]
	b1 := messageBlock[i*4+1]
	b2 := messageBlock[i*4+2]
	b3 := messageBlock[i*4+3]

	return bytesToUint32(b0, b1, b2, b3)
}

func bytesToUint32(b0, b1, b2, b3 byte) uint32 {
	return uint32(b0)<<24 | uint32(b1)<<16 | uint32(b2)<<8 | uint32(b3)
}

// parseMessage parses a message into 512-bit sized chunks
// for SHA-256 as defined in section 5.2.1 of FIPS 180-4
func parseMessage(message []byte) [][]byte {
	var chunks [][]byte

	chunkSize := 512 / 8

	for i := 0; i < len(message); i += chunkSize {
		end := i + chunkSize
		chunks = append(chunks, message[i:end])
	}

	return chunks
}

func padMessage(message []byte) []byte {
	padded := message
	padded = append(padded, 0b10000000)
	// Padded message is: message plus 1 '1' bit, plus N '0' bits plus (message as big endian 64 bit int)
	// N grows in order to ensure that the padded message ends on a 512 bit (64 byte) boundary (required for the hash function)
	zeroBytes := calculateZeroBits(len(message)) / 8
	padded = append(padded, bytes.Repeat([]byte{0x00}, zeroBytes)...)
	padded = binary.BigEndian.AppendUint64(padded, uint64(len(message)*8))

	return padded
}

func fractionalPartOfCubeRoot(n int) uint32 {
	res := math.Pow(float64(n), 1.0/3)
	_, fractional := math.Modf(res)
	scaled := fractional * (1 << 32)
	return uint32(scaled)
}

func fractionalPartOfSquareRoot(n int) uint32 {
	res := math.Sqrt(float64(n))
	_, fractional := math.Modf(res)
	scaled := fractional * (1 << 32)
	return uint32(scaled)
}

func getPrimeNumber(n int) int {
	primes := []int{}
	i := 1
next:
	for {
		i++

		if len(primes) == n {
			return primes[n-1]
		}

		// Iterate primes checking if each prime divides i
		for j := 0; j < len(primes); j++ {
			p := primes[j]
			if p > i {
				continue
			}
			if i%p == 0 {
				goto next
			}
		}

		primes = append(primes, i)
	}
}

// initialHashValues returns the initial values hash values
// for SHA-256 as defined in section 5.3.3 of FIPS 180-4
func initialHashValues() []uint32 {
	ihv := make([]uint32, 8, 8)
	for i := 0; i < 8; i++ {
		ihv[i] = fractionalPartOfSquareRoot(getPrimeNumber(i + 1))
	}

	return ihv
}

// kConstant returns the K_i constant for SHA-256 as defined
// in section 4.2.2 of FIPS 180-4
func kConstant(i int) uint32 {
	return fractionalPartOfCubeRoot(getPrimeNumber(i + 1))
}

// calculateZeroBits implements section 5.1.1 of FIPS 180-4
func calculateZeroBits(lengthInBytes int) int {
	lengthInBits := lengthInBytes * 8
	// Each block is 512 bits. We can effectively discard each full block.
	// We are required to pad a '1' bit.
	bitsConsumedInLastBlock := (lengthInBits + 1) % 512

	var zeros int

	// A 64 bit int is required, so we need enough space for it
	if bitsConsumedInLastBlock < 448 {
		// We need to pad from bitsRemainingInLastBlock up to 448
		zeros = 448 - bitsConsumedInLastBlock
	} else {
		// If there is not enough room for the 64 bit int,
		// we require another block. e.g if the message is 1000 bits long,
		// then a 64 bit int won't fit in the 1024 bits from 2 blocks.

		// 960 = 448 + 512
		zeros = 960 - bitsConsumedInLastBlock
	}

	paddedMessageLength := lengthInBits + 1 + zeros + 64
	if paddedMessageLength%512 != 0 {
		panic(fmt.Sprintf("Padded message is %d bits long which is not a multiple of 512!", paddedMessageLength))
	}

	return zeros
}

// Choice function
func ch(x, y, z uint32) uint32 {
	return (x & y) ^ (^x & z)
}

// Majority function
func maj(x, y, z uint32) uint32 {
	return (x & y) ^ (x & z) ^ (y & z)
}

func upperSigma0(x uint32) uint32 {
	return rotr(2, x) ^ rotr(13, x) ^ rotr(22, x)
}

func upperSigma1(x uint32) uint32 {
	return rotr(6, x) ^ rotr(11, x) ^ rotr(25, x)
}

func lowerSigma0(x uint32) uint32 {
	return rotr(7, x) ^ rotr(18, x) ^ shr(3, x)
}

func lowerSigma1(x uint32) uint32 {
	return rotr(17, x) ^ rotr(19, x) ^ shr(10, x)
}

func rotr(n int, x uint32) uint32 {
	return (x >> n) | (x << (32 - n))
}

func shr(n int, x uint32) uint32 {
	return x >> n
}
