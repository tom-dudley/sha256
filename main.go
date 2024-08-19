package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func main() {
	fmt.Println("vim-go")
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
