package network

import (
	"authex/helpers"
	"authex/model"
	"authex/network/abi"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

// NodeClient is the client to interact with the ethereum node
type NodeClient struct {
	keystore *keystore.KeyStore
	signer   accounts.Account
	client   *ethclient.Client
	wsURL    string
	// contracts
	accessControl *abi.AccessControl
	// when a new token need to be monitored is sent to this
	// channel, the client will start monitoring it
	Tokens chan string
	// track the currently monitored tokens
	monitoredTokens map[string]int
	// channel to send monitored transfers
	Transfers chan *model.BalanceChange
}

// NewNodeClient create a new node client
func NewNodeClient(settings *model.Settings, transfers chan *model.BalanceChange) (*NodeClient, error) {
	// check if the account has the admin privileges
	client, err := ethclient.Dial(settings.Network.RPCEndpoint)
	if err != nil {
		return nil, err
	}

	// open the keystore
	ks := keystore.NewKeyStore(settings.Identity.KeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	signer, err := helpers.UnlockAccount(ks, settings.Identity.SignerAddress, settings.Identity.Password)
	if err != nil {
		return nil, err
	}

	// Get the access control contract
	address := common.HexToAddress(settings.Identity.AccessContractAddress)
	ac, err := abi.NewAccessControl(address, client)
	if err != nil {
		return nil, err
	}

	return &NodeClient{
		keystore:        ks,
		client:          client,
		wsURL:           settings.Network.WSEndpoint,
		signer:          signer,
		accessControl:   ac,
		monitoredTokens: map[string]int{},
		Tokens:          make(chan string),
		Transfers:       transfers,
	}, nil
}

// Run begin listening for network events
func (n *NodeClient) Run() {
	monitors := 0
	for {
		token, ok := <-n.Tokens
		if !ok {
			log.Infof("token monitor channel closed")
		}
		monitors++
		log.Infof("received token to monitor: %s", token)
		// check if the token is already monitored
		if _, ok := n.monitoredTokens[token]; ok {
			log.Infof("token already monitored: %s", token)
			continue
		}
		// start monitoring the token
		go n.monitorToken(token, monitors)
		n.monitoredTokens[token] = monitors
	}
}

// monitorToken listen to transfer events for the given erc20 token
// to and from the CLOB address. This way the CLOB can track balance changes
// of the users
func (n *NodeClient) monitorToken(address string, monitorID int) {
	client, err := ethclient.Dial(n.wsURL)
	if err != nil {
		log.Errorf("[monitor: %d] websocket connection: %v", monitorID, err)
		return
	}
	contractAddress := common.HexToAddress(address)
	logs := make(chan *abi.ERC20Transfer)

	erc20, err := abi.NewERC20(contractAddress, client)
	if err != nil {
		log.Errorf("[monitor: %d] erc20 contract: %v", monitorID, err)
		return
	}

	// REMEMBER! this is the balance on the CLOB, not of the wallet

	sub, err := erc20.WatchTransfer(&bind.WatchOpts{}, logs, nil, nil)
	defer sub.Unsubscribe()

	if err != nil {
		log.Errorf("[monitor: %d] logs subscription filter: %v", monitorID, err)
		return
	}

	for {
		select {
		case err := <-sub.Err():
			log.Errorf("[monitor: %d] log: %v", monitorID, err)
		case t := <-logs:

			log.Infof("[monitor: %d] transfer: %v", monitorID, t)
			if helpers.IsZeroAddress(t.From) {
				continue
			}
			if helpers.IsZeroAddress(t.To) {
				continue
			}

			var deltas []*model.BalanceDelta
			if t.From.Hex() == n.signer.Address.Hex() {
				d := decimal.NewFromBigInt(t.Value.Neg(t.Value), 0)
				// it's a withdrawal
				deltas = append(deltas, model.NewBalanceDelta(t.From.Hex(), d))
			} else {
				// is a deposit
				d := decimal.NewFromBigInt(t.Value, 0)
				deltas = append(deltas, model.NewBalanceDelta(t.To.Hex(), d))
			}
			n.Transfers <- &model.BalanceChange{
				TokenAddress: address,
				BlockNumber:  t.Raw.BlockNumber,
				Deltas:       deltas,
			}
		}
	}
}

// GetSigner return the address of the signer account for this server instance
func (n *NodeClient) GetSigner() string {
	return n.signer.Address.Hex()
}

// IsERC20 check if the given address is an ERC20 token
func (n *NodeClient) IsERC20(address string) (bool, error) {
	contractAddress := common.HexToAddress(address)
	erc20, err := abi.NewERC20(contractAddress, n.client)
	if err != nil {
		return false, err
	}
	return erc20 != nil, nil
}

// IsAdmin check if the given address is admin
// an address is admin if it has the DEFAULTADMINROLE role in the access control contract
func (n *NodeClient) IsAdmin(address string) (bool, error) {
	role, err := n.accessControl.DEFAULTADMINROLE(nil)
	if err != nil {
		return false, err
	}
	return n.accessControl.HasRole(nil, role, common.HexToAddress(address))
}

// ExecuteWithdraw transfer the given amount of tokens to the given address
func (n *NodeClient) ExecuteWithdraw(tokenAddress string, amount *big.Int, to string) (string, error) {
	contractAddress := common.HexToAddress(tokenAddress)
	erc20, err := abi.NewERC20(contractAddress, n.client)
	if err != nil {
		return "", err
	}
	tx, err := erc20.Transfer(&bind.TransactOpts{}, common.HexToAddress(to), amount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// Setup import the keyfile in the local keystore and return the address
func Setup(settings *model.Settings) error {
	err := os.RemoveAll(settings.Identity.KeystorePath)
	if err != nil {
		return err
	}

	nc, err := NewNodeClient(settings, nil)
	if err != nil {
		return err
	}
	signer := nc.GetSigner()
	isAdmin, err := nc.IsAdmin(signer)
	if err != nil {
		return err
	}
	if !isAdmin {
		err := fmt.Errorf("account %s is not admin", signer)
		return err
	}
	return nil
}
