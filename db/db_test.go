package db_test

import (
	"authex/clob"
	"authex/db"
	"authex/model"
	"context"
	_ "embed"
	"fmt"

	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	// set the log level to debug
	log.SetLevel(log.DEBUG)
}

func createContainer(ctx context.Context) (testcontainers.Container, string, error) {

	var (
		dbName = "test_db"
		dbUser = "test_user"
		dbPass = "test_password"
	)
	var env = map[string]string{
		"POSTGRES_PASSWORD": dbPass,
		"POSTGRES_USER":     dbUser,
		"POSTGRES_DB":       dbName,
	}
	var port = "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
			Cmd:          []string{"-c", "fsync=off"},
		},
		Started:      true,
		ProviderType: testcontainers.ProviderPodman,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, "", fmt.Errorf("failed to start container: %w", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, "", fmt.Errorf("failed to get container external port: %w", err)
	}

	log.Infof("postgres container ready and running at port: ", p.Port())

	time.Sleep(time.Second)

	dbURI := fmt.Sprint("postgres://", dbUser, ":", dbPass, "@localhost:", p.Port(), "/", dbName, "?sslmode=disable")

	db, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return container, dbURI, fmt.Errorf("failed to establish database connection: %w", err)
	}
	db.Close()
	return container, dbURI, nil
}

var settings model.Settings

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, uri, err := createContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}

	settings = model.Settings{
		DB: struct {
			URI            string
			MaxConnections int
		}{
			URI: uri,
		},
	}
	exitStatus := m.Run()
	if err = container.Terminate(ctx); err != nil {
		log.Fatalf("error terminating container %v", err)
	}
	cancel()
	os.Exit(exitStatus)
}

func TestConnection_SaveMarket(t *testing.T) {
	dbCli, err := db.NewConnection(&settings)
	assert.NoError(t, err, "error connecting to the database")
	err = dbCli.InitializeSchema()
	assert.NoError(t, err, "error initializing the database")

	type args struct {
		marketAddress string
		base          *model.Asset
		quote         *model.Asset
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "ok",
			wantErr: nil,
			args: args{
				marketAddress: "0x1",
				base: &model.Asset{
					Symbol:  "USD",
					Address: "usd",
					Class:   model.AssetOffChain,
				},
				quote: &model.Asset{
					Symbol:  "EUR",
					Address: "eur",
					Class:   model.AssetOffChain,
				},
			},
		},
		{
			name:    "ERR: duplicated",
			wantErr: db.ErrInsert,
			args: args{
				marketAddress: "0x1",
				base: &model.Asset{
					Symbol:  "USD",
					Address: "usd",
					Class:   model.AssetOffChain,
				},
				quote: &model.Asset{
					Symbol:  "EUR",
					Address: "eur",
					Class:   model.AssetOffChain,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = dbCli.SaveMarket(tt.args.marketAddress, tt.args.base, tt.args.quote)
			assert.ErrorIs(t, err, tt.wantErr)
			_, err = dbCli.GetMarketByAddress(tt.args.marketAddress)
			assert.NoError(t, err, "market must exists")

		})
	}
}

func TestConnection_handleMatch(t *testing.T) {

	var (
		_alice   = "0xaa992902d88EA6192585B72D0B01C020F036bb99"
		_bob     = "0xbbD65e1115Ff895b6c0F313ca050A613a150c940"
		_usd_eur = "0xd36cfda1a6607e8b79d0c9ea784346a6e21fad86"
		_usd     = "0x505f49beeda8b41a13274e3622c64e61d087a796"
		_eur     = "0x60c197cc20da7f7d7c4d019fb9e66cd79b223c6c"
	)

	dbCli, err := db.NewConnection(&settings)
	assert.NoError(t, err, "error connecting to the database")
	err = dbCli.InitializeSchema()
	assert.NoError(t, err, "error initializing the database")
	// start the db
	go dbCli.Run()
	clob := clob.NewPool(dbCli.Matches)
	go clob.Run()

	type balance struct {
		accountAddress string
		assetAddress   string
		balance        decimal.Decimal
	}
	type price struct {
		marketAddress string
		price         decimal.Decimal
	}
	type args struct {
		initialBalances []balance
		orders          []*model.SignedRequest[model.Order]
		market          model.MarketInfo
	}
	tests := []struct {
		name         string
		args         args
		wantBalances []balance
		wantPrice    []price
		wantErr      error
	}{
		{
			name: "ok",
			args: args{
				initialBalances: []balance{
					{
						accountAddress: _alice,
						assetAddress:   _usd,
						balance:        decimal.NewFromInt(1_000_000),
					},
					{
						accountAddress: _bob,
						assetAddress:   _eur,
						balance:        decimal.NewFromInt(1_000_000),
					},
				},
				market: model.MarketInfo{
					Address: _usd_eur,
					Base: model.Asset{
						Symbol:  "USD",
						Address: _usd,
						Class:   model.AssetOffChain,
					},
					Quote: model.Asset{
						Symbol:  "EUR",
						Address: _eur,
						Class:   model.AssetOffChain,
					},
				},
				orders: []*model.SignedRequest[model.Order]{
					{
						Payload: model.Order{
							// alice buys 1 eur(quote) for 100 usd(base)
							Market: _usd_eur,
							Price:  "100",
							Size:   1,
							Side:   model.SideBid,
						},
						From: _alice,
					},
					{
						Payload: model.Order{
							// bob sells 1 eur(quote) for 100 usd(base)
							Market: _usd_eur,
							Price:  "100",
							Size:   1,
							Side:   model.SideAsk,
						},
						From: _bob,
					},
				},
			},
			wantBalances: []balance{
				{
					accountAddress: _alice,
					assetAddress:   _usd,
					balance:        decimal.NewFromInt(1_000_000).Sub(decimal.NewFromInt(100)),
				},
				{
					accountAddress: _alice,
					assetAddress:   _eur,
					balance:        decimal.NewFromInt(1),
				},
				{
					accountAddress: _bob,
					assetAddress:   _usd,
					balance:        decimal.NewFromInt(100),
				},
				{
					accountAddress: _bob,
					assetAddress:   _eur,
					balance:        decimal.NewFromInt(1_000_000).Sub(decimal.NewFromInt(1)),
				},
			},
			wantPrice: []price{
				{
					marketAddress: _usd_eur,
					price:         decimal.NewFromInt(100),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create the market
			err = dbCli.SaveMarket(tt.args.market.Address, &tt.args.market.Base, &tt.args.market.Quote)
			assert.NoError(t, err, "error saving market")

			// create the initial balances
			for _, balance := range tt.args.initialBalances {
				err = dbCli.UpdateBalance(balance.accountAddress, balance.assetAddress, balance.balance)
				assert.NoError(t, err, "error saving balance")
			}
			// make sure that the balances are correct
			for _, balance := range tt.args.initialBalances {
				balanceAmount, err := dbCli.GetBalance(balance.accountAddress, balance.assetAddress)
				assert.NoError(t, err, "error getting balance")
				assert.Equal(t, balance.balance, balanceAmount, "balance must match")
			}

			for _, order := range tt.args.orders {
				// TODO incorrect
				zeroQuote := decimal.NewFromInt(0)
				dbCli.ValidateOrder(&order.Payload, order.From, zeroQuote)
				clob.Inbound <- order
			}
			// give the clob some time to process the orders
			time.Sleep(50 * time.Millisecond)

			for _, wantBalance := range tt.wantBalances {
				balance, err := dbCli.GetBalance(wantBalance.accountAddress, wantBalance.assetAddress)
				assert.NoError(t, err, "error getting balance")
				assert.Equalf(t, wantBalance.balance, balance, "balance mismatch account: %s, asset: %s, balance: %s", wantBalance.accountAddress, wantBalance.assetAddress, balance)
			}

			// for _, wantPrice := range tt.wantPrice {
			// 	price, err := dbCli.GetMarketPrice(wantPrice.marketAddress)
			// 	assert.NoError(t, err, "error getting price")
			// 	assert.Equal(t, wantPrice.price, price, "price must match")
			// }
		})
	}
	// sleep for 50ms to allow the db to close
	time.Sleep(50 * time.Millisecond)
	dbCli.Close()
	clob.Close()
}
