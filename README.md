## Utility to determine Tally votes for a wallet over a set period
  - Queries all wallet address transactions that have interacted with the Tally Hop contract 0xed8Bdb5895B8B7f9Fdb3C087628FD8410E853D48
  - Extracts transaction ID, gas fee, date, Tally proposal ID, vote reason, and vote, for each transaction
  - Requirements:
      - [Go](https://go.dev/dl/)
      - [Blockdaemon Free API Suite key](https://app.blockdaemon.com/signin/register) to query indexed data


### Usage
```shell
go run main.go <apikey> <walletaddress> <from-epoch> <to-epoch>
```

### Example
```shell
go run main.go zpka_... 0xbE5C59873f34-redacted-d72F5F330 1693526400 1712809148
   txId: 0xd81fb35df06-redacted-1f1406e1a847d792028e09
   gas: 0.00044754978765335204 ETH
   date: 2023-10-09 16:24:47 -0700 PDT
   proposalId: 9784707434912981-redacted-6768683443315035555006717324695
   reason: I am voting For treasury diversification of Hop DAO to cater to cash outlays, de-risk market movements, and promote sustainability.
   supported: 1
```