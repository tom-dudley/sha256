package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func main() {
	fmt.Println("TODO")
}

func padMessage(message []byte) []byte {
	padded := message
	padded = append(padded, 0b10000000)
	// Padded message is: message plus 1 '1' bit, plus N '0' bits plus (message as big endian 64 bit int)
	// N grows in order to ensure that the padded message ends on a 512 bit (64 byte) boundary (required for the hash function)
	zeroBytes := 64 - 1 - 8 - len(message)
	padded = append(padded, bytes.Repeat([]byte{'0'}, zeroBytes)...)
	padded = binary.BigEndian.AppendUint64(padded, uint64(len(message)))

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
	fmt.Printf("%d bits consumed in last block\n", bitsConsumedInLastBlock)

	var zeros int

	// A 64 bit int is required, so we need enough space for it
	if bitsConsumedInLastBlock < 448 {
		// We need to pad from bitsRemainingInLastBlock up to 448
		zeros = 448 - bitsConsumedInLastBlock
		fmt.Println("More than 64 bits remaining")
		fmt.Printf("Zeroes: %d\n", zeros)
	} else {
		// If there is not enough room for the 64 bit int,
		// we require another block. e.g if the message is 1000 bits long,
		// then a 64 bit int won't fit in the 1024 bits from 2 blocks.

		// 960 = 448 + 512
		zeros = 960 - bitsConsumedInLastBlock + 1
		fmt.Println("Less than 64 bits remaining")
		fmt.Printf("Zeroes: %d\n", zeros)
	}

	paddedMessageLength := lengthInBits + 1 + zeros + 64
	if paddedMessageLength%512 != 0 {
		panic(fmt.Sprintf("Padded message is %d bits long which is not a multiple of 512!", paddedMessageLength))
	}

	return zeros
}

// Key steps:
// - Preprocess
//   - Pad the message
//   - Parse the message
// - Hash
