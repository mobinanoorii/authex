package helpers

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func ComputeMarketAddress(baseAddress string, quoteAddress string) (string, error) {
	address := []string{baseAddress, quoteAddress}
	sort.Strings(address)
	var market []byte
	for _, a := range address {
		b, err := hex.DecodeString(strings.ToLower(strings.TrimPrefix(a, "0x")))
		if err != nil {
			return "", err
		}
		market = append(market, b...)
	}
	h := crypto.Keccak256(market)
	return fmt.Sprint("0x", hex.EncodeToString(h[12:])), nil
}

func ParseAmount(value string) (decimal.Decimal, error) {
	b, err := decimal.NewFromString(value)
	if err != nil {
		return b, err
	}
	return b, nil
}

// IID generates a a unique incident id to be used in error logs
// and that is returned to the user in the error response
func IID() string {
	return uuid.New().String()
}
