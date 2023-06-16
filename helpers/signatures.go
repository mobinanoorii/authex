package helpers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/term"
)

func Sign(keystorePath, address string, msg any) (signature string, err error) {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	var signer accounts.Account
	// get the account
	if len(ks.Accounts()) == 0 {
		err = fmt.Errorf("no accounts found in the keystore")
		return
	}
	if len(ks.Accounts()) == 1 {
		signer = ks.Accounts()[0]
	} else {
		for _, acc := range ks.Accounts() {
			if acc.Address.String() == address {
				signer = acc
				break
			}
		}
		if signer.Address.String() == "" {
			err = fmt.Errorf("no account found for address %s", address)
			return
		}
	}
	// unlock the account
	pass := PasswordPrompt(signer.Address.String())
	if err = ks.Unlock(signer, pass); err != nil {
		fmt.Println("error unlocking account:", err)
		return
	}
	// prepare the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	msgBytes = crypto.Keccak256(msgBytes)
	// sign the message
	sigBytes, err := ks.SignHash(signer, msgBytes)
	if err != nil {
		return
	}
	signature = hex.EncodeToString(sigBytes)
	return
}

// PasswordPrompt asks for a string value using the label.
// The entered value will not be displayed on the screen
// while typing.
func PasswordPrompt(account string) string {
	label := fmt.Sprintf("Enter password for account %s:", account)
	var s string
	for {
		fmt.Fprint(os.Stderr, label+" ")
		b, _ := term.ReadPassword(int(syscall.Stdin))
		s = string(b)
		if s != "" {
			break
		}
	}
	fmt.Println()
	return s
}