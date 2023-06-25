# AutHEx

Ethereum Centralized Exchange, a Central Limit Order Book with attitude.

> THE PROJECT IS IN VERY EARLY STAGE

[![codecov](https://codecov.io/gh/noandrea/authex/branch/main/graph/badge.svg?token=8ZZXN3XOAJ)](https://codecov.io/gh/noandrea/authex)

The project binaries contain both the server and the client, the client is a command line tool that can be used to interact with the server.

## Usage

To run the server use:

```console
authex server start
```

### Deployment

An example of Kubernetes manifests for deployment can be found in `deploy` folder.


### Setup

Authex uses postgresql as persistence storage, to start the system make sure you have access to a
postgresql instance and run the following commands to create the target database

```console
sudo -u postgres psql
psql (14.8 (Ubuntu 14.8-0ubuntu0.22.04.1))
Type "help" for help.

postgres=#
```

once you are in the postgresql terminal use the following commands to create a database and a role that
will be used by the AutoCLOB application.

```
postgres=# create role app with password 'app';
CREATE ROLE
postgres=# alter role app with login;
ALTER ROLE
postgres=# create database authex owner app;
CREATE DATABASE
postgres=# quit
```

Once the database is created run the setup command:

```console
authex server setup -d postgres://app:app@localhost:5432/authex
```


## Endpoints

The server exposes the following endpoints. Note that all the requests made to the server need to be signed using your account private key.

### Administration endpoints


| Method | Path                  | Help                                                                |
| ------ | --------------------- | ------------------------------------------------------------------- |
| POST   | /admin/markets        | Register a new market (requires admin privileges)                   |
| POST   | /admin/accounts/fund  | Fund an account (requires admin privileges)                         |
| POST   | /admin/accounts/allow | Add an account to the allowed list (requires admin privileges)      |
| POST   | /admin/accounts/block | Remove an account from the allowed list (requires admin privileges) |


A client is provided to interact with the server, to use it run the following command:

```console
authex admin
Group of admin commands

Usage:
  authex admin [command]

Available Commands:
  fund            Fund an account with an asset (modify the account balance in AutHEx)
  register-market Register a new market

Flags:
      --from string            the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore
  -h, --help                   help for admin
  -k, --keystore-path string   Path to the keystore directory (default "./_private/keystore")
  -n, --non-interactive        commands will not prompt for input (password)
  -w, --password string        the password to unlock the sender account
  -e, --rest-url string        the base URL of the REST API (default "http://127.0.0.1:2306")

Additional help topics:
  authex admin authorize       Authorize a new account to trade

Use "authex admin [command] --help" for more information about a command.

```



### Query endpoints

| Method | Path                                      | Help                    |
| ------ | ----------------------------------------- | ----------------------- |
| GET    | /query/markets                            | Get all markets         |
| GET    | /query/markets/:address                   | Get a market by address |
| GET    | /query/markets/:address/quote/:side/:size | Get a market quote      |
| GET    | /query/orders/:id                         | Get an order by id      |

A client is provided to interact with the server, to use it run the following command:

```console
authex query
Group of query commands

Usage:
  authex query [command]

Available Commands:
  market      Query a market
  markets     Get all markets
  order       Query an order
  quote       Get a quote for a market

Flags:
  -h, --help              help for query
  -e, --rest-url string   the base URL of the REST API (default "http://127.0.0.1:2306")

Use "authex query [command] --help" for more information about a command.

```

### Account endpoints

| Method | Path                              | Help                                       |
| ------ | --------------------------------- | ------------------------------------------ |
| POST   | /account/orders                   | Post a new buy or sell order               |
| POST   | /account/orders/cancel            | Cancel an order                            |
| POST   | /account/withdraw                 | Withdraw funds from the CLOB               |
| GET    | /account/orders/:id               | Get an order by id                         |
| GET    | /account/:address/orders          | Get all orders for an account              |
| GET    | /account/:address/balance/:symbol | Get the balance of an account for a symbol |

A client is provided to interact with the server, to use it run the following command:

```console
authex account
Group of user commands

Usage:
  authex account [command]

Available Commands:
  ask          Submit a new order
  ask-market   Submit a new market order
  bid          Submit a new buy limit order
  bid-market   Submit a new buy limit order
  cancel-order Cancel an order
  withdraw     Withdraw tokens from the exchange.

Flags:
      --from string            the address to send the transaction from (must be an account in the keystore), only required when there is more than one account in the keystore
  -h, --help                   help for account
  -k, --keystore-path string   Path to the keystore directory (default "./_private/keystore")
  -n, --non-interactive        commands will not prompt for input (password)
  -w, --password string        the password to unlock the sender account
  -e, --rest-url string        the base URL of the REST API (default "http://127.0.0.1:2306")

Use "authex account [command] --help" for more information about a command.
```

## Binaries

Binaries are available for Linux on the [release page](https://github.com/noandrea/authex/releases).

Artifacts checksums are available in the `checksums.txt` file, to verify the integrity of the binaries you can use the following command:

```console

export AUTHEX_VERSION=0.5.0
curl -LO https://github.com/noandrea/authex/releases/download/v"$AUTHEX_VERSION"/authex_"$AUTHEX_VERSION"_checksums.txt
curl -LO https://github.com/noandrea/authex/releases/download/v"$AUTHEX_VERSION"/authex_"$AUTHEX_VERSION"_linux_amd64.tar.gz
test "$(cat authex_"$AUTHEX_VERSION"_checksums.txt| grep linux_amd64.tar.gz)" = "$(sha256sum authex_"$AUTHEX_VERSION"_linux_amd64.tar.gz)" && echo Yes || echo No

```

The checksum is signed with the following PGP key at [839409EF054B9E6A](https://keys.openpgp.org/search?q=839409EF054B9E6A). To verify the signature you can use the following command:

```console
export AUTHEX_VERSION=0.5.0
gpg --keyserver keys.openpgp.org --recv 839409EF054B9E6A
curl -LO https://github.com/noandrea/authex/releases/download/v"$AUTHEX_VERSION"/authex_"$AUTHEX_VERSION"_checksums.txt.sig
gpg --verify authex_"$AUTHEX_VERSION"_checksums.txt.sig
```

the output should be similar to the following:

```console
gpg: assuming signed data in 'authex_0.5.0_checksums.txt'
gpg: Signature made Fri 23 Jun 2023 07:02:29 PM CEST
gpg:                using EDDSA key C1956FA64BA3D28EF0D23643839409EF054B9E6A
gpg: Good signature from "Andrea Giacobino (Releases signature key) <no.andrea@gmail.com>" [ultimate]
```