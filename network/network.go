package network

import (
	"authex/model"
	"authex/network/abi"
	_ "embed"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NodeClient struct {
	keystore *keystore.KeyStore
	signer   accounts.Account
	client   *ethclient.Client
	wsClient *ethclient.Client
	// contracts
	accessControl *abi.AccessControl
	token         *abi.ERC20
}

func NewNodeClient(settings *model.Settings) (*NodeClient, error) {
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
	// get the erc20 contract
	address = common.HexToAddress(settings.Identity.TokenAddress)
	erc20, err := abi.NewERC20(address, client)
	if err != nil {
		return nil, err
	}

	return &NodeClient{
		keystore: ks,
		client:   client,
		signer:   ks.Accounts()[0],
		// contracts
		accessControl: ac,
		token:         erc20,
	}, nil
}

// Listen listen for network events and update balances
func (p *NodeClient) Run() {
	// for {
	// 	select {
	// 	case order := <-p.Inbound:
	// 		p.handleOrder(order)
	// 	default:
	// 		// channel is closed
	// 		return
	// 	}
	// }
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

	nc, err := NewNodeClient(settings)
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
