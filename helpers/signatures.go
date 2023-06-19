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

func Sign(keystorePath, address, password string, promptPassword bool, msg any) (signature string, err error) {
	if promptPassword {
		password = PasswordPrompt(address)
	}
	return doSign(keystorePath, address, password, msg)
}

func doSign(keystorePath, address, password string, msg any) (signature string, err error) {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	var signer accounts.Account
	// get the account
	if len(ks.Accounts()) == 0 {
		err = fmt.Errorf("no accounts found in the keystore")
		return
	}
	if len(ks.Accounts()) == 1 {
		signer = ks.Accounts()[0]
		if signer.Address.String() != address {
			err = fmt.Errorf("address %s does not match the keystore account", address)
			return
		}
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
	if err = ks.Unlock(signer, password); err != nil {
		println("error unlocking account:", err)
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
		b, _ := term.ReadPassword(syscall.Stdin)
		s = string(b)
		break
	}
	println()
	return s
}
