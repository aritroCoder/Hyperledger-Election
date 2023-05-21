/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	abac "github.com/aritroCoder/decentralized-election/smart-contract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	abacSmartContract, err := contractapi.NewChaincode(&abac.SmartContract{})
	if err != nil {
		log.Panicf("Error creating abac chaincode: %v", err)
	}

	if err := abacSmartContract.Start(); err != nil {
		log.Panicf("Error starting abac chaincode: %v", err)
	}
}
