package main

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func messageLength(message string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint64(len(message)))
	return buf.Bytes()
}

// Each block is made up of 512 bits, i.e. 64 bytes.
// The message has a '1' bit appended and then '0's up until
// 64 bits from the end of the last block, then append the length
// of the message as a a big endian 64 bit int (8 bytes).
func TestPadMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  []byte
		expected []byte
	}{
		{
			// 8 bytes for message length plus one for the '1', plus the message = 14 bytes
			// Therefore we require 64 - 14 bytes of zeros, i.e. 50
			name:    "convert string less than 512 bits",
			message: []byte("Hello"),
			expected: append(
				[]byte{'H', 'e', 'l', 'l', 'o', 0b10000000},
				append(
					bytes.Repeat([]byte{'0'}, 64-8-1-len("Hello")),
					messageLength("Hello")...,
				)...,
			),
		},
		{
			name:    "convert another string less than 512 bits",
			message: []byte("Another message"),
			expected: append(
				[]byte{'A', 'n', 'o', 't', 'h', 'e', 'r', ' ', 'm', 'e', 's', 's', 'a', 'g', 'e', 0b10000000},
				append(
					bytes.Repeat([]byte{'0'}, 64-8-1-len("Another message")),
					messageLength("Another message")...,
				)...,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, 64, len(test.expected))
			paddedMessage := padMessage(test.message)
			require.Equal(t, test.expected, paddedMessage)
		})
	}
}

func TestFractionalPartOfCubeRoot(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected uint32
	}{
		{
			name:     "2",
			number:   2,
			expected: uint32(0x428a2f98),
		},
		{
			name:     "3",
			number:   3,
			expected: uint32(0x71374491),
		},
		{
			name:     "64th prime",
			number:   getPrimeNumber(64),
			expected: uint32(0xc67178f2),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fractionalPart := fractionalPartOfCubeRoot(test.number)
			require.Equal(t, test.expected, fractionalPart)
		})
	}
}

func TestFractionalPartOfSquare(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected uint32
	}{
		{
			name:     "2",
			number:   2,
			expected: uint32(0x6a09e667),
		},
		{
			name:     "3",
			number:   3,
			expected: uint32(0xbb67ae85),
		},
		{
			name:     "8th prime",
			number:   getPrimeNumber(8),
			expected: uint32(0x5be0cd19),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fractionalPart := fractionalPartOfSquareRoot(test.number)
			require.Equal(t, test.expected, fractionalPart)
		})
	}
}

func TestGetPrimeNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected int
	}{
		{
			name:     "Get 2nd prime number",
			number:   2,
			expected: 3,
		},
		{
			name:     "Get 10th prime number",
			number:   10,
			expected: 29,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			primeNumber := getPrimeNumber(test.number)
			require.Equal(t, test.expected, primeNumber)
		})
	}
}
