package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
