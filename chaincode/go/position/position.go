package main


import (
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("PositionChaincode")

const indexName = `Position`

type PositionChaincode struct {
}

// Position is the main data type stored in ledger
type Position struct {
	Balance      	Balance `json:"balance"`
	Security        string 	`json:"security"`
	Quantity        int 	`json:"quantity"`
}

type Balance struct {
	Account 		string 	`json:"account"`
	Division 		string 	`json:"division"`
}

// required for history
type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     PositionValue  `json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

// required for history
type PositionValue struct {
	Quantity        int 	`json:"quantity"`
}

// **** Position Methods **** //
func (this *Position) toStringArray() ([]string) {
	return []string{
		this.Balance.Account,
		this.Balance.Division,
		this.Security,
	}
}

func (this *Position) toCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(indexName, this.toStringArray())
}

func (this *Position) existsIn(stub shim.ChaincodeStubInterface) (bool) {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return false
	}

	bytes, _ := stub.GetState(compositeKey)
	if bytes == nil {
		return false
	}

	return true
}

func (this *Position) loadFrom(stub shim.ChaincodeStubInterface) (error) {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return err
	}

	bytes, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return this.fillFromLedgerValue(bytes)
}

func (this *Position) upsertIn(stub shim.ChaincodeStubInterface) (error) {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return err
	}

	value, err := this.toLedgerValue()
	if err != nil {
		return err
	}

	err = stub.PutState(compositeKey, value)
	if err != nil {
		return err
	}

	return nil
}

func (this *Position) fillFromCompositeKeyParts(compositeKeyParts []string) (error) {
	if len(compositeKeyParts) != 3 {
		return errors.New("Composite key parts array length must be 3.")
	}

	this.Balance.Account 	= compositeKeyParts[0]
	this.Balance.Division 	= compositeKeyParts[1]
	this.Security 			= compositeKeyParts[2]

	return nil
}

func (this *Position) fillFromArgs(args []string) (error) {
	if len(args) != 4 {
		return errors.New("Incorrect number of arguments. Expecting 4.")
	}

	this.Balance.Account 	= args[0]
	this.Balance.Division 	= args[1]
	this.Security 			= args[2]

	quantity, err := strconv.Atoi(args[3])
	if err != nil {
		return errors.New("cannot convert to quantity")
	}
	this.Quantity = quantity

	return nil
}

func (this *Position) toLedgerValue() ([]byte, error) {
	return json.Marshal([]string{strconv.Itoa(this.Quantity)})
}

func (this *Position) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Position) fillFromLedgerValue(bytes []byte) (error) {
	var str []string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	quantity, err := strconv.Atoi(str[0])
	if err != nil {
		logger.Error("cannot convert to quantity", err)
	}
	this.Quantity = quantity

	return nil
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
		"receive, transfer, query, history, status. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func (t *PositionChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	position := Position{}
	err := position.fillFromArgs(args)
	if err != nil {
		//TODO change from 500 to bad request 400
		return shim.Error(err.Error())
	}

	if position.upsertIn(stub) != nil {
		return shim.Error("Position upsertIn error.")
	}

	return shim.Success([]byte("Position updated."))
}

func (t *PositionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	positions := []Position{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		position := Position{}

		err = position.fillFromLedgerValue(response.Value)
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = position.fillFromCompositeKeyParts(compositeKeyParts)
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
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo, " +
			"security, quantity, reference, positionDate, tradeDate")
	}

	position := Position{}
	err := position.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := position.toCompositeKey(stub)
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

// **** main method **** //
func main() {
	err := shim.Start(new(PositionChaincode))
	if err != nil {
		logger.Errorf("Error starting Position chaincode: %s", err)
	}
}

