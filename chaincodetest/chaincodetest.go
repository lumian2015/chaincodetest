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
	Hospital   string
	Disease    string
	Link       string
	Checkpoint bool
}

type Token struct {
	Owner        string //will change to certificate later
	Availability bool
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
		value, err := stub.GetState(args[0] + "token")
		if err != nil {
			return nil, errors.New("Put failed")
		}
		token := Token{}
		if value != nil {
			err = json.Unmarshal([]byte(value), &token)
			if token.Owner != args[1] {
				return nil, errors.New("Put failed")
			}
		} else {
			return nil, errors.New("Put failed")
		}
		patient_id := args[0]
		record := Record{}
		record.Hospital = args[1]
		record.Disease = args[2]
		record.Link = args[3]
		record.Checkpoint, err = strconv.ParseBool(args[4])
		if err != nil {
			return nil, errors.New("Failed to change checkpoint from string to bool")
		}
		value, err = stub.GetState(patient_id)
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
			return nil, fmt.Errorf("Encoding failed %s", err)
		} else {
			err = stub.PutState(patient_id, []byte(b))
			if err != nil {
				fmt.Printf("Error putting state %s", err)
				return nil, fmt.Errorf("put operation failed. Error updating state: %s", err)
			}
		}
		token.Availability = true
		t, err := json.Marshal(token)
		stub.PutState(patient_id+"token", []byte(t))
		successed_msg := "Put success"
		stub.SetEvent("successedEvent", []byte(successed_msg))
		return nil, nil
	case "getToken":
		if len(args) != 2 {
			return nil, errors.New("getToken operation must provide the key and the hospital's name")
		}
		patient_id := args[0] + "token"
		value, err := stub.GetState(patient_id)
		if err != nil {
			return nil, errors.New("Get token failed")
		}
		token := Token{}
		if value != nil {
			json.Unmarshal([]byte(value), &token)
			if token.Availability == false && token.Owner != args[1] {
				return nil, errors.New("Get token failed")
			}
		}
		token.Owner = args[1]
		token.Availability = false
		b, err := json.Marshal(token)
		if err != nil {
			return nil, fmt.Errorf("Encoding failed %s", err)
		} else {
			err = stub.PutState(patient_id, []byte(b))
			if err != nil {
				return nil, fmt.Errorf("getToken operation failed. Error updating state: %s", err)
			}
		}
		successed_msg := args[1]
		stub.SetEvent(args[0], []byte(successed_msg))
		return nil, nil
	default:
		return nil, errors.New("Unsupported operation")
	}
}

// Query has two functions
// get - takes one argument, a key, and returns the value for the key
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {

	case "get":
		if len(args) < 1 {
			return nil, errors.New("get operation must include one argument, a key")
		}
		value, err := stub.GetState(args[0] + "token")
		if err != nil {
			fmt.Printf("Failed to get the patient's token")
			return nil, errors.New("failed to get the patient's token")
		}
		token := Token{}
		if value != nil {
			err = json.Unmarshal([]byte(value), &token)
			if token.Owner != args[0] && token.Availability == false {
				fmt.Printf("Someone is writing the data, try later")
				return nil, errors.New("Someone is writing the data, try later")
			}
		}
		key := args[0]
		value, err = stub.GetState(key)
		if err != nil {
			return nil, fmt.Errorf("get operation failed. Error accessing state: %s", err)
		}
		if value == nil {
			return nil, fmt.Errorf("no related data has been stored")
		}
		fmt.Printf("the result is %s\n", string(value))
		return value, nil

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
