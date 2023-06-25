package db_test

import (
	"authex/db"
	"authex/model"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DbName = "test_db"
	DbUser = "test_user"
	DbPass = "test_password"
)

func createContainer(ctx context.Context) (testcontainers.Container, string, error) {

	var env = map[string]string{
		"POSTGRES_PASSWORD": DbPass,
		"POSTGRES_USER":     DbUser,
		"POSTGRES_DB":       DbName,
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
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, "", fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, "", fmt.Errorf("failed to get container external port: %v", err)
	}

	log.Println("postgres container ready and running at port: ", p.Port())

	time.Sleep(time.Second)

	dbURI := fmt.Sprint("postgres://", DbUser, ":", DbPass, "@localhost:", p.Port(), "/", DbName, "?sslmode=disable")

	db, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return container, dbURI, fmt.Errorf("failed to establish database connection: %v", err)
	}
	db.Close()
	return container, dbURI, nil
}

var dbURI string

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, uri, err := createContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}
	cancel()
	dbURI = uri
	defer func() {
		container.Terminate(ctx)
	}()
	os.Exit(m.Run())
}

func TestConnection_SaveMarket(t *testing.T) {

	opt := &model.Settings{
		DB: struct {
			URI            string
			MaxConnections int
		}{
			URI: dbURI,
		},
	}

	dbCli, err := db.NewConnection(opt)
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
			err := dbCli.SaveMarket(tt.args.marketAddress, tt.args.base, tt.args.quote)
			assert.ErrorIs(t, err, tt.wantErr)
			_, err = dbCli.GetMarketByAddress(tt.args.marketAddress)
			assert.NoError(t, err, "market must exists")

		})
	}
}
