package helpers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// IsEmpty check if a string is empty
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

const ZeroAddress = "0x0000000000000000000000000000000000000000"

// Check if the address is a zero address
func IsZeroAddress(address common.Address) bool {
	return strings.TrimSpace(address.Hex()) == ZeroAddress
}

// Hash hashes a string using keccak256
func Hash(f string, o ...string) string {
	r := crypto.Keccak256([]byte(f))
	for _, v := range o {
		r = append(r, crypto.Keccak256([]byte(v))...)
	}
	return hex.EncodeToString(r)
}

// AsAddress given a string, return a valid ethereum address
// this is useful to generate a valid address from a asset name
func AsAddress(input string) string {
	h := crypto.Keccak256([]byte(strings.ToLower(input)))
	return "0x" + hex.EncodeToString(h[12:])
}

func ComputeMarketID(baseAddress string, quoteAddress string) string {
	address := []string{baseAddress, quoteAddress}
	sort.Strings(address)
	var market []byte
	for _, a := range address {
		// TODO handle error
		b, _ := hex.DecodeString(a)
		market = append(market, b...)
	}
	h := crypto.Keccak256(market)
	return "0x" + hex.EncodeToString(h[12:])
}

func ParseAmount(value string) (*big.Int, error) {
	b, ok := new(big.Int).SetString(value, 10)
	if !ok {
		err := fmt.Errorf("cannot parse value %s to an amount", value)
		return nil, err
	}
	return b, nil
}
