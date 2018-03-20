package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

type Record struct {
	Disease    string
	Timestamp  int
	Link       string
	Checkpoint bool
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// Invoke has two functions
// put - takes five arguements, a key and five values, and stores them in the state
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {
	case "put":
		if len(args) < 5 {
			return nil, errors.New("put operation must include five arguments, a key and four values")
		}
		patient_id := args[0]
		record := Record{}
		record.Disease = args[1]
		timestamp, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, errors.New("Failed to change timestamp from string to int")
		}
		record.Timestamp = timestamp
		record.Link = args[3]
		record.Checkpoint, err = strconv.ParseBool(args[4])
		if err != nil {
			return nil, errors.New("Failed to change checkpoint from string to bool")
		}
		value, err := stub.GetState(patient_id)
		if err != nil {
			return nil, errors.New("Failed to get the patient's records")
		}
		var res []Record
		if value == nil {
			res = append(res, record)
		} else {
			err = json.Unmarshal([]byte(value), &res)
			if err != nil {
				fmt.Printf("decoding failed %s", err)
				return nil, errors.New("decoding failed")
			}
			res = append(res, record)
		}
		b, err := json.Marshal(res)
		if err != nil {
			fmt.Println("encoding failed")
		} else {
			err = stub.PutState(patient_id, []byte(b))
			if err != nil {
				fmt.Printf("Error putting state %s", err)
				return nil, fmt.Errorf("put operation failed. Error updating state: %s", err)
			}
		}
		return nil, nil

	default:
		return nil, errors.New("Unsupported operation")
	}
}

// Query has two functions
// get - takes one argument, a key, and returns the value for the key
// keys - returns all keys stored in this chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {

	case "get":
		if len(args) < 1 {
			return nil, errors.New("get operation must include one argument, a key")
		}
		key := args[0]
		value, err := stub.GetState(key)
		if err != nil {
			return nil, fmt.Errorf("get operation failed. Error accessing state: %s", err)
		}
		if value == nil {
			return nil, fmt.Errorf("no related data has been stored")
		}
		fmt.Printf("the result is %s\n", string(value))
		return value, nil

	case "keys":

		keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []string
		for keysIter.HasNext() {
			key, _, iterErr := keysIter.Next()
			if iterErr != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}
			keys = append(keys, key)
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil

	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
