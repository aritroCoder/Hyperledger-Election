package abac

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type ElectionSystem struct {
	IsVoted      map[string]bool `json:"isvoted"`
	Candidates   map[string]int  `json:"candidates"`
	TopCandidate string          `json:"topcandidate"`
	IsVoteOn     bool            `json:"isvoteon"`
	Voters       map[string]int  `json:"voters"`
	VoterCount   int             `json:"votercount"`
}

// initialize world state
func (s *SmartContract) InitializeLedger(ctx contractapi.TransactionContextInterface) error {
	electionStruct := ElectionSystem{
		IsVoted:      make(map[string]bool),
		Candidates:   make(map[string]int),
		TopCandidate: "",
		IsVoteOn:     false,
		Voters:       make(map[string]int),
		VoterCount:   0,
	}

	// put data back to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState("1", newElectionJson)
}

// add voters to vote
func (s *SmartContract) AddVoter(ctx contractapi.TransactionContextInterface) error {
	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return err
	}
	if electionJSON == nil {
		return fmt.Errorf("world state not initialized")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return err
	}

	if !electionStruct.IsVoteOn {
		return fmt.Errorf("Vote is not going on")
	}

	if _, ok := electionStruct.Voters[clientID]; ok {
		return fmt.Errorf("you are already added as a voter")
	}

	electionStruct.Voters[clientID] = electionStruct.VoterCount + 1
	electionStruct.VoterCount++

	// put data back to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState("1", newElectionJson)
}

// edited as per https://stackoverflow.com/questions/69637688/json-encode-in-golang-fabric-chaincode-behavior
type Persons struct {
	Dumb bool `json:"dumb"`
	clientId string `json:"clientid,omitempty" metadata:",optional"`
	numId    int `json:"numid,omitempty" metadata:",optional"`
}

// display all voters with their ids
func (s *SmartContract) GetAllVoters(ctx contractapi.TransactionContextInterface) ([]*Persons, error) {
	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return nil, err
	}
	if electionJSON == nil {
		return nil, fmt.Errorf("world state not initialized")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return nil, err
	}

	if !electionStruct.IsVoteOn {
		return nil, fmt.Errorf("Vote is not going on")
	}

	var voterList []*Persons
	for k, v := range electionStruct.Voters {
		newItem := Persons{clientId: k, numId: v}
		voterList = append(voterList, &newItem)
	}
	
	return voterList, nil
}

// display all candidates with their ids
func (s *SmartContract) GetAllCandidates(ctx contractapi.TransactionContextInterface) ([]*Persons, error) {
	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return nil, err
	}
	if electionJSON == nil {
		return nil, fmt.Errorf("world state not initialized")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return nil, err
	}

	if !electionStruct.IsVoteOn {
		return nil, fmt.Errorf("Vote is not going on")
	}

	var voterList []*Persons
	for k := range electionStruct.Candidates {
		newItem := Persons{clientId: k, numId: electionStruct.Voters[k]}
		voterList = append(voterList, &newItem)
	}

	return voterList, nil
}

// add a candidate to the list
func (s *SmartContract) AddCandidate(ctx contractapi.TransactionContextInterface) (string, error) {
	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return "", err
	}

	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return "", err
	}
	if electionJSON == nil {
		return "", fmt.Errorf("world state not initialized")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return "", err
	}

	if !electionStruct.IsVoteOn {
		return "", fmt.Errorf("Vote is not going on")
	}

	// if person not exists in voters list, add them
	if _, ok := electionStruct.Voters[clientID]; !ok {
		s.AddVoter(ctx)
	}

	electionStruct.Candidates[clientID] = 0 // start with zero votes

	// put data back to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return "", err
	}
	return clientID, ctx.GetStub().PutState("1", newElectionJson)
}

func (s *SmartContract) Vote(ctx contractapi.TransactionContextInterface, id string) error {
	// Get ID of submitting client identity and fetch data from world state
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return err
	}
	if electionJSON == nil {
		return fmt.Errorf("does not exist")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return err
	}
	//finish getting data from world state

	if !electionStruct.IsVoteOn {
		return fmt.Errorf("Vote is not going on")
	}

	if electionStruct.IsVoted[clientID] {
		return fmt.Errorf("you have already voted")
	}

	electionStruct.Candidates[id]++
	electionStruct.IsVoted[clientID] = true

	// update data to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState("1", newElectionJson)
}

func (s *SmartContract) PublishResults(ctx contractapi.TransactionContextInterface) (string, error) {
	// fetch data from world state
	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return "", err
	}
	if electionJSON == nil {
		return "", fmt.Errorf("does not exist")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return "", err
	}
	//finish getting data from world state

	if !electionStruct.IsVoteOn {
		return "", fmt.Errorf("Vote is not going on")
	}

	//find the highest voted candidate using map
	var topCandidate string
	var topVotes int = 0
	for k, v := range electionStruct.Candidates {
		if topVotes < v {
			topCandidate = k
			topVotes = v
		}
	}

	electionStruct.IsVoteOn = false

	// update data to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState("1", newElectionJson)
	if err != nil {
		return "", err
	}

	return topCandidate, nil
}

func (s *SmartContract) SetVoteOn(ctx contractapi.TransactionContextInterface) error {
	// fetch data from world state
	electionJSON, err := ctx.GetStub().GetState("1")
	if err != nil {
		return err
	}
	if electionJSON == nil {
		return fmt.Errorf("does not exist")
	}

	var electionStruct ElectionSystem
	err = json.Unmarshal(electionJSON, &electionStruct)
	if err != nil {
		return err
	}
	// fetched data

	electionStruct.IsVoteOn = true

	// update data to world state
	newElectionJson, err := json.Marshal(electionStruct)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState("1", newElectionJson)
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

