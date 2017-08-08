package main


import (
	"fmt"
	"encoding/json"
	"time"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("BookChaincode")

const indexName = `Book`

// BookChaincode
type BookChaincode struct {
}

type BookValue struct {
	Quantity   		int 	`json:"quantity"`
	Deponent 		string 	`json:"deponent"`
}

type Balance struct {
	Account 		string 	`json:"account"`
	Division 		string 	`json:"division"`
}

type Book struct {
	Balance 		Balance `json:"balance"`
	Security        string 	`json:"security"`
	Quantity   		int 	`json:"quantity"`
	Deponent	 	string  `json:"deponent"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     BookValue  		`json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

func (t *BookChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### BookChaincode Init ###########")

	_, args := stub.GetFunctionAndParameters()

	return t.init(stub, args)
}

func (t *BookChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### BookChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "init" {
		return t.init(stub, args)
	}
	if function == "move" {
		return t.move(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}
	if function == "history" {
		return t.history(stub, args)
	}

	err := fmt.Sprintf("Unknown function, check the first argument, must be one of: " +
		"move, query, history. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func (t *BookChaincode) init(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// deponent, account, division, security, quantity
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting deponent, account, division, security, quantity")
	}

	deponent := args[0]
	account := args[1]
	division := args[2]
	security := args[3]
	quantity, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error(err.Error())
	}

	// account-division-security
	key, err := stub.CreateCompositeKey(indexName, []string{account, division, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(BookValue{Deponent: deponent, Quantity: quantity})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BookChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// accountFrom, divisionFrom, accountTo, divisionTo, security, quantity
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting accountFrom, divisionFrom, accountTo, divisionTo, security, quantity")
	}

	accountFrom := args[0]
	divisionFrom := args[1]
	accountTo := args[2]
	divisionTo := args[3]
	security := args[4]
	quantity, err := strconv.Atoi(args[5])
	if err != nil {
		return shim.Error(err.Error())
	}

	keyFrom, err := stub.CreateCompositeKey(indexName, []string{accountFrom, divisionFrom, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := stub.GetState(keyFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	if bytes == nil {
		return shim.Error("cannot find balance")
	}

	var valueFrom BookValue
	err = json.Unmarshal(bytes, &valueFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	if valueFrom.Quantity < quantity {
		return shim.Error("cannot move quantity less than current balance")
	}

	valueFrom.Quantity = valueFrom.Quantity - quantity

	newBytes, err := json.Marshal(valueFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(keyFrom, newBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	keyTo, err := stub.CreateCompositeKey(indexName, []string{accountTo, divisionTo, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err = stub.GetState(keyTo)
	if err != nil {
		return shim.Error(err.Error())
	}

	if bytes == nil {
		newBytes, err = json.Marshal(BookValue{Quantity: quantity})
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		var valueTo BookValue
		err = json.Unmarshal(bytes, &valueTo)
		if err != nil {
			return shim.Error(err.Error())
		}

		valueTo.Quantity = valueTo.Quantity + quantity

		newBytes, err = json.Marshal(valueTo)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	err = stub.PutState(keyTo, newBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BookChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	books := []Book{}
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

		var value BookValue
		err = json.Unmarshal(responseRange.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		book := Book {
			Balance: Balance {
				Account: compositeKeyParts[0],
				Division: compositeKeyParts[1],
			},
			Security: compositeKeyParts[2],
			Quantity: value.Quantity,
			Deponent: value.Deponent,
		}

		books = append(books, book)
	}

	result, err := json.Marshal(books)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *BookChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting account, division, security")
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
	err := shim.Start(new(BookChaincode))
	if err != nil {
		logger.Errorf("Error starting Book chaincode: %s", err)
	}
}

