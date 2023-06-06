package network

import (
	"authex/model"
	"authex/network/abi"
	_ "embed"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/gommon/log"
)

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
	Transfers chan *model.ERC20Transfer
}

func NewNodeClient(settings *model.Settings, transfers chan *model.ERC20Transfer) (*NodeClient, error) {
	// check if the account has the admin privileges
	client, err := ethclient.Dial(settings.Network.RPCEndpoint)
	if err != nil {
		return nil, err
	}
	// open the keystore
	ks := keystore.NewKeyStore(settings.Identity.KeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(ks.Accounts()) == 0 {
		jsonBytes, err := os.ReadFile(settings.Identity.KeyFile)
		if err != nil {
			return nil, err
		}
		_, err = ks.Import(jsonBytes, settings.Identity.Password, settings.Identity.Password)
		if err != nil {
			return nil, err
		}
	}
	// Get the access control contract
	address := common.HexToAddress(settings.Identity.AccessContractAddress)
	ac, err := abi.NewAccessControl(address, client)
	if err != nil {
		return nil, err
	}

	return &NodeClient{
		keystore:      ks,
		client:        client,
		wsURL:         settings.Network.WSEndpoint,
		signer:        ks.Accounts()[0],
		accessControl: ac,
		Tokens:        make(chan string),
		Transfers:     transfers,
	}, nil
}

// Listen listen for network events and update balances
func (p *NodeClient) Run() {

	monitors := 0
	for {
		select {
		case token := <-p.Tokens:
			monitors++
			log.Infof("received token to monitor: %s", token)
			// check if the token is already monitored
			if _, ok := p.monitoredTokens[token]; ok {
				log.Infof("token already monitored: %s", token)
				continue
			}
			// start monitoring the token
			go p.monitorToken(token, monitors)
			p.monitoredTokens[token] = monitors
		default:
			// channel is closed
			return
		}
	}
}

// monitorToken listen to transfer events for the given erc20 token
func (nc *NodeClient) monitorToken(address string, monitorID int) {
	client, err := ethclient.Dial(nc.wsURL)
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
			// send the transfer to the channel
			nc.Transfers <- &model.ERC20Transfer{
				From:         t.From.Hex(),
				To:           t.To.Hex(),
				Amount:       t.Value,
				TokenAddress: address,
				BlockNumber:  t.Raw.BlockNumber,
			}
		}
	}
}

// GetSigner return the address of the signer account for this server instance
func (nc *NodeClient) GetSigner() string {
	return nc.signer.Address.Hex()
}

func (nc *NodeClient) IsAdmin(address string) (bool, error) {
	role, err := nc.accessControl.DEFAULTADMINROLE(nil)
	if err != nil {
		return false, err
	}
	return nc.accessControl.HasRole(nil, role, common.HexToAddress(address))
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

/*

bytes32 public constant DEFAULT_ADMIN_ROLE = 0x00;
function hasRole(bytes32 role, address account) public view virtual override returns (bool) {
    return _roles[role].members[account];
}
*/
