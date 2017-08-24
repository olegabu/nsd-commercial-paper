package main


import (
	"fmt"
	"encoding/json"
	"time"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/olegabu/nsd-commercial-paper-common"
)

var logger = shim.NewLogger("BookChaincode")

const bookIndex = `Book`

// BookChaincode
type BookChaincode struct {
}

type BookValue struct {
	Quantity   		int 	`json:"quantity"`
}


//TODO reuse Position struct
type Book struct {
	Balance 		nsd.Balance `json:"balance"`
	Security        string 	`json:"security"`
	Quantity   		int 	`json:"quantity"`
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

	type bookInit struct {
		Account     string `json:"account"`
		Division    string `json:"division"`
		Security    string `json:"security"`
		Quantity    string `json:"quantity"`
	}

	var bookInits []bookInit
	if err := json.Unmarshal([]byte(args[0]), &bookInits); err == nil && len(bookInits) != 0 {
		for _, entry := range bookInits {
			t.put(stub, []string{entry.Account, entry.Division, entry.Security, entry.Quantity})
		}
	}

	return shim.Success(nil)
}

func (t *BookChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### BookChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "put" {
		return t.put(stub, args)
	}
	if function == "move" {
		return t.move(stub, args)
	}
	if function == "check" {
		return t.check(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}
	if function == "history" {
		return t.history(stub, args)
	}

	err := fmt.Sprintf("Unknown function, check the first argument, must be one of: " +
		"put, move, check, query, history. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func (t *BookChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// account, division, security, quantity
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting account, division, security, quantity")
	}

	account := args[0]
	division := args[1]
	security := args[2]
	quantity, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error(err.Error())
	}

	// account-division-security
	key, err := stub.CreateCompositeKey(bookIndex, []string{account, division, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(BookValue{Quantity: quantity})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BookChaincode) check(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// account, division, security, quantity
	if len(args) != 4 {
		return pb.Response{Status:400, Message: "Incorrect number of arguments. " +
			"Expecting account, division, security, quantity"}
	}

	account := args[0]
	division := args[1]
	security := args[2]
	quantity, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error(err.Error())
	}

	keyFrom, err := stub.CreateCompositeKey(bookIndex, []string{account, division, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := stub.GetState(keyFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	if bytes == nil {
		return pb.Response{Status:404, Message: "cannot find position"}
	}

	var value BookValue
	err = json.Unmarshal(bytes, &value)
	if err != nil {
		return shim.Error(err.Error())
	}

	if value.Quantity < quantity {
		return pb.Response{Status:409, Message: "quantity less than current balance"}
	}

	return shim.Success(nil)
}

func (t *BookChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	//check stored list for this instructions has been executed already
	if instruction.ExistsIn(stub) {
		if err := instruction.LoadFrom(stub); err != nil {
			return pb.Response{Status: 500, Message: "Instruction cannot be loaded."}
		}

		if instruction.Value.Status == nsd.InstructionExecuted {
			return pb.Response{Status: 409, Message: "Already executed."}
		}
	}

	accountFrom := instruction.Key.Transferer.Account
	divisionFrom := instruction.Key.Transferer.Division
	security := instruction.Key.Security
	quantity, _ := strconv.Atoi(instruction.Key.Quantity)
	accountTo := instruction.Key.Receiver.Account
	divisionTo := instruction.Key.Receiver.Division

	keyFrom, err := stub.CreateCompositeKey(bookIndex, []string{accountFrom, divisionFrom, security})
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := stub.GetState(keyFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	if bytes == nil {
		return pb.Response{Status:404, Message: "cannot find position"}
	}

	var valueFrom BookValue
	err = json.Unmarshal(bytes, &valueFrom)
	if err != nil {
		return shim.Error(err.Error())
	}

	if valueFrom.Quantity < quantity {
		return pb.Response{Status:409, Message: "cannot move quantity less than current balance"}
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

	keyTo, err := stub.CreateCompositeKey(bookIndex, []string{accountTo, divisionTo, security})
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

	if instruction != (nsd.Instruction{}) {
		instruction.Value.Status = nsd.InstructionExecuted

		//save to the ledger list of executed instructions
		if err := instruction.UpsertIn(stub); err != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}

		if err := instruction.EmitState(stub); err != nil {
			return pb.Response{Status: 500, Message: "Event emission failure."}
		}
	}

	return shim.Success(nil)
}


func (t *BookChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(bookIndex, []string{})
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
			Balance: nsd.Balance {
				Account: compositeKeyParts[0],
				Division: compositeKeyParts[1],
			},
			Security: compositeKeyParts[2],
			Quantity: value.Quantity,
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
		return pb.Response{Status:400, Message: "Incorrect number of arguments. " +
			"Expecting account, division, security"}
	}

	//account-division-security
	key, err := stub.CreateCompositeKey(bookIndex, args)
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

