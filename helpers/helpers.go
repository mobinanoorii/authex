package helpers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
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

var (
	ErrInput = errors.New("invalid input")
)

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
		if IsEmpty(a) {
			return "", errors.Join(ErrInput, errors.New("empty addresses not allowed"))
		}
		b, err := hex.DecodeString(strings.ToLower(strings.TrimPrefix(a, "0x")))
		if err != nil {
			return "", errors.Join(ErrInput, err)
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

func EnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func EnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return strings.ToLower(value) == "true"
	}
	return fallback
}
