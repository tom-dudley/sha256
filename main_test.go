package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
// TODO: These tests are currently failing on a last byte mismatch
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
					bytes.Repeat([]byte{0x00}, calculateZeroBits(len("Hello"))/8),
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
					bytes.Repeat([]byte{0x00}, calculateZeroBits(len("Another message"))/8),
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

func TestCalculateZeroBits(t *testing.T) {
	tests := []struct {
		name          string
		messageLength int
		expected      int
	}{
		{
			name:          "3 bytes",
			messageLength: 3,
			expected:      423,
		},
		{
			name:          "1000 bytes",
			messageLength: 1000,
			expected:      127,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, calculateZeroBits(test.messageLength))
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

func TestBytesToUint32(t *testing.T) {
	b0 := byte(0x12)
	b1 := byte(0x34)
	b2 := byte(0x56)
	b3 := byte(0x78)

	require.Equal(t, uint32(0x12345678), bytesToUint32(b0, b1, b2, b3))
}

func TestRotr(t *testing.T) {
	x := uint32(0x4d2c6ea2)
	shift := 4
	expected := uint32(0x24d2c6ea)

	actual := rotr(shift, x)
	require.Equal(t, expected, actual)
}

func TestGetWordFromMessageBlock(t *testing.T) {
	message := []byte{0x4b, 0xbb, 0x12, 0x6f, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	paddedMessage := padMessage(message)
	messageBlocks := parseMessage(paddedMessage)
	messageBlock := messageBlocks[0]

	require.Equal(t, bytesToUint32(0x4b, 0xbb, 0x12, 0x6f), getWordFromMessageBlock(messageBlock, 0))
	require.Equal(t, bytesToUint32(0xaa, 0xbb, 0xcc, 0xdd), getWordFromMessageBlock(messageBlock, 1))
	require.Equal(t, bytesToUint32(0xee, 0xff, 0x80, 0x00), getWordFromMessageBlock(messageBlock, 2))
}

func TestSha256Hash(t *testing.T) {
	message := []byte("It works!")
	hash := sha256.Sum256(message)
	hashHex := hex.EncodeToString(hash[:])

	require.Equal(t, hashHex, sha256Digest(message))
}
