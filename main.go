package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"log"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Response struct {
	Total int `json:"total"`
	Data  []struct {
		ID      string `json:"id"`
		BlockID string `json:"block_id"`
		Date    int    `json:"date"`
		Status  string `json:"status"`
		Meta    struct {
			To string `json:"to"`
		} `json:"meta"`
		Events []struct {
			Amount int64 `json:"amount"`
		} `json:"events"`
	} `json:"data"`
}

func DecodeTransactionInputData(contractABI *abi.ABI, data []byte) (string, string, uint8) {
	methodSigData := data[:4]
	inputsSigData := data[4:]
	method, err := contractABI.MethodById(methodSigData)
	if err != nil {
		log.Fatal(err)
	}
	inputsMap := make(map[string]interface{})
	if err := method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
		log.Fatal(err)
	}
	proposalId := inputsMap["proposalId"].(*big.Int).Text(10)
	return proposalId, inputsMap["reason"].(string), inputsMap["support"].(uint8)
}

const abiJSON = `[
    {
        "constant": false,
        "inputs": [
            {
                "name": "proposalId",
                "type": "uint256"
            },
            {
                "name": "support",
                "type": "uint8"
            },
            {
                "name": "reason",
                "type": "string"
            }
        ],
        "name": "castVoteWithReason",
        "outputs": [],
        "payable": false,
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

const tallyHopContract = "0xed8Bdb5895B8B7f9Fdb3C087628FD8410E853D48"

func getCommandLineArgs() (string, string, string, string) {
	apikey := ""
	walletaddress := ""
	from := ""
	to := ""

	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go <apikey> <walletaddress> <from-epoch> <to-epoch>")
		os.Exit(1)
	}

	apikey = os.Args[1]
	walletaddress = os.Args[2]
	from = os.Args[3]
	to = os.Args[4]

	return apikey, walletaddress, from, to
}

func main() {
	apikey, walletaddress, from, to := getCommandLineArgs()

	// Define Tally Vote contract ABI
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// ToDo paginate through all requests
	url := "https://svc.blockdaemon.com/universal/v1/ethereum/mainnet/account/" + walletaddress + "/txs?from=" + from + "&to=" + to + "&order=asc&page_size=100"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-Key", apikey)

	resp, _ := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making the HTTP request:", err)
	}
	defer resp.Body.Close()

	// Create go-ethereum/rpc client with header-based authentication
	rpcClient, err := rpc.Dial("https://svc.blockdaemon.com/ethereum/mainnet/native")
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client:", err)
	}
	rpcClient.SetHeader("X-API-KEY", apikey)
	client := ethclient.NewClient(rpcClient)

	var response Response
	json.NewDecoder(resp.Body).Decode(&response)
	for _, tx := range response.Data {
		if tx.Meta.To == tallyHopContract {
			fmt.Println("txId:", tx.ID)
			fmt.Println("gas:", float64(tx.Events[0].Amount)/1000000000/1000000000, "ETH") // ! ToDo determine gas amount in USD at the time of the transaction
			fmt.Println("date:", time.Unix(int64(tx.Date), 0))

			// decode the raw transaction data.input to determine the proposal ID, vote, and reason
			txn, _, _ := client.TransactionByHash(context.Background(), common.HexToHash(tx.ID))
			proposalId, reason, support := DecodeTransactionInputData(&contractABI, txn.Data())
			fmt.Println("proposalId:", proposalId)
			fmt.Println("reason:", reason)
			fmt.Println("supported:", support, "\n")
		}
	}
}
