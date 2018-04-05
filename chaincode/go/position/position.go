package main


import (
	"fmt"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/Altoros/nsd-commercial-paper-common"
)

var logger = shim.NewLogger("PositionChaincode")

type PositionChaincode struct {
}

// required for history
type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     PositionValue  	`json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

// required for history
type PositionValue struct {
	Quantity        string 	`json:"quantity"`
}

// **** Chaincode Methods **** //
func (t *PositionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### PositionChaincode Init ###########")

	return shim.Success(nil)
}

func (t *PositionChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### PositionChaincode Invoke ###########")

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
		"put, query, history. But got: %v", function)
	logger.Error(err)
	return shim.Error(err)
}

func (t *PositionChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	position := nsd.Position{}
	err := position.FillFromArgs(args)
	if err != nil {
		//TODO change from 500 to bad request 400
		return shim.Error(err.Error())
	}

	if position.UpsertIn(stub) != nil {
		return shim.Error("Position upsertIn error.")
	}

	return shim.Success([]byte("Position updated."))
}

func (t *PositionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(nsd.PositionIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	positions := []nsd.Position{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		position := nsd.Position{}

		err = position.FillFromLedgerValue(response.Value)
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = position.FillFromCompositeKeyParts(compositeKeyParts)
		if err != nil {
			return shim.Error(err.Error())
		}

		positions = append(positions, position)

	}

	result, err := json.Marshal(positions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *PositionChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting account, division, security")
	}

	position := nsd.Position{}
	err := position.FillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := position.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	it, err := stub.GetHistoryForKey(compositeKey)
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


		var values []string
		err = json.Unmarshal(response.GetValue(), &values)
		if err != nil {
			return shim.Error(err.Error())
		}
		if len(values) > 0 {
			entry.Value.Quantity = values[0]
		}

		modifications = append(modifications, entry)
	}

	result, err := json.Marshal(modifications)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

// **** main method **** //
func main() {
	err := shim.Start(new(PositionChaincode))
	if err != nil {
		logger.Errorf("Error starting Position chaincode: %s", err)
	}
}

