package main


import (
	"fmt"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("SecurityChaincode")

const indexName = `Security`

// SecurityChaincode
type SecurityChaincode struct {
}

type SecurityValue struct {
	Status      	string 	`json:"status"`
}

type Security struct {
	Security        string 	`json:"security"`
	Status      	string 	`json:"status"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     SecurityValue  	`json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

func (t *SecurityChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### SecurityChaincode Init ###########")

	_, args := stub.GetFunctionAndParameters()

	return t.put(stub, args)
}

func (t *SecurityChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### SecurityChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "put" {
		return t.put(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}
	if function == "history" {
		return t.history(stub, args)
	}

	err := fmt.Sprintf("Unknown function, check the first argument, must be one of: " +
		"put, query, history. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func (t *SecurityChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// security, status
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security, status")
	}

	security := args[0]
	status := args[1]

	key, err := stub.CreateCompositeKey(indexName, []string{security})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(SecurityValue{Status: status})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SecurityChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	securitys := []Security{}
	for it.HasNext() {
		responseRange, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		//account-division-security
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		var value SecurityValue
		err = json.Unmarshal(responseRange.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		security := Security {
			Security: compositeKeyParts[0],
			Status: value.Status,
		}

		securitys = append(securitys, security)
	}

	result, err := json.Marshal(securitys)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *SecurityChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security")
	}

	//account-division-security
	key, err := stub.CreateCompositeKey(indexName, args)
	if err != nil {
		return shim.Error(err.Error())
	}

	it, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	modifications := []KeyModificationValue{}

	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var entry KeyModificationValue

		entry.TxId = response.GetTxId()
		entry.IsDelete = response.GetIsDelete()
		ts := response.GetTimestamp()

		if ts != nil {
			entry.Timestamp = time.Unix(ts.Seconds, int64(ts.Nanos)).String()
		}

		err = json.Unmarshal(response.GetValue(), &entry.Value)
		if err != nil {
			return shim.Error(err.Error())
		}

		modifications = append(modifications, entry)
	}

	result, err := json.Marshal(modifications)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(SecurityChaincode))
	if err != nil {
		logger.Errorf("Error starting Security chaincode: %s", err)
	}
}

