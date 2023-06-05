# Contracts go bindings 


To genereate contracts abi bindings for go use the [`abigen`](https://geth.ethereum.org/docs/tools/abigen) tool. 


The sources are generated with the following commands

#### ERC20 

The ERC20 binding is generated using the ABI from the openzeppelin implementation: `openzeppelin-contracts@4.8.1/ERC20.sol`

```console
abigen --abi ERC20.abi --pkg abi --type ERC20 --out ERC20.go
```

#### AccessControl 

The AccessControl binding is generated using the ABI from the openzeppelin implementation `openzeppelin-contracts@4.8.1/AccessControl.sol`

```console
abigen --abi AccessControl.abi --pkg abi --type AccessControl --out AccessControl.go
```
