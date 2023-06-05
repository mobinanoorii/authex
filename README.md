# AutHEx

Autonity Hybrid Exchange, a Central Limit Order Book with attitude.

> THE PROJECT IS IN VERY EARLY STAGE

## Usage

To run the server use:

```console
authex start --help
TODO
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
authex setup -d postgres://app:app@localhost:5432/authex
```


### Endpoints

The server exposes the following endpoints. Note that all the requests made to the server need to be signed using your account private key. 

#### Submit an order

Submit an order to the CLOB 

```curl
➜ curl -v -H "Content-Type: application/json" \
    -d '{"signature":"abcdefg", order: {price: ""}}' \
    http://localhost:2306/orders/submit | jq
*   Trying 127.0.0.1:2306...
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
```


