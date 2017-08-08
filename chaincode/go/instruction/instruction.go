package main


import (
	"fmt"
	//"strconv"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"time"

//	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("InstructionChaincode")

// 9 fields we're matching on
const indexName = `accountFrom-divisionFrom-accountTo-divisionTo-security-quantity-reference-instructionDate-tradeDate`

// InstructionChaincode
type InstructionChaincode struct {
}

type InstructionValue struct {
	DeponentFrom 	string 	`json:"deponentFrom"`
	DeponentTo		string 	`json:"deponentTo"`
	Status      	string 	`json:"status"`
	Initiator 		string 	`json:"initiator"`
}

type Source struct {
	Account 	string 	`json:"account"`
	Division 	string 	`json:"division"`
}

type Reason struct {
	Document 	string 	`json:"document"`
	Description string 	`json:"description"`
	DocumentDate 	string 	`json:"documentDate"`
}

type Instruction struct {
	Transferer 		Source 	`json:"transferer"`
	Receiver   		Source 	`json:"receiver"`
	Security   		string 	`json:"security"`
	Quantity   		string 	`json:"quantity"`
	Reference  		string 	`json:"reference"`
	InstructionDate string 	`json:"instructionDate"`
	TradeDate  		string 	`json:"tradeDate"`
	DeponentFrom	string  `json:"deponentFrom"`
	DeponentTo		string  `json:"deponentTo"`
	Status     		string 	`json:"status"`
	Initiator 		string 	`json:"initiator"`
	Reason 			Reason 	`json:"reason"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     InstructionValue  `json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

func (t *InstructionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### InstructionChaincode Init ###########")

	return shim.Success(nil)
}

func (t *InstructionChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### InstructionChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "receive" {
		return t.receive(stub, args)
	}
	if function == "transfer" {
		return t.transfer(stub, args)
	}
	if function == "status" {
		return t.status(stub, args)
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

func (t *InstructionChaincode) GetOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	data := certificate[strings.Index(string(certificate), "-----"):strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]

	return organization
}

func (t *InstructionChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// deponentFrom, accountFrom, divisionFrom, *, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting deponentFrom, accountFrom, divisionFrom, accountTo, divisionTo, security, quantity, reference, " +
			"instructionDate, tradeDate, reason")
	}

	deponentFrom := args[0]
	accountFrom := args[1]
	divisionFrom := args[2]
	deponentTo := t.GetOrganization(stub)
	accountTo := args[3]
	divisionTo := args[4]
	security := args[5]
	quantity := args[6] //quantity, _ := strconv.Atoi(args[2])
	reference := args[7]
	instructionDate := args[8]
	tradeDate := args[9] // := time.Now().UTC().Unix()
	//reason := args[10]

	//accountFrom-divisionFrom-accountTo-divisionTo-security-quantity-reference-instructionDate-tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{accountFrom, divisionFrom,
		accountTo, divisionTo, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{DeponentFrom: deponentFrom, DeponentTo: deponentTo,
		Status: "initiated", Initiator: "receiver"})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(key))
}

func (t *InstructionChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// *, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting accountFrom, divisionFrom, deponentTo, accountTo, divisionTo, security, quantity, reference, " +
			"instructionDate, tradeDate, reason.")
	}

	deponentFrom := t.GetOrganization(stub)
	accountFrom := args[0]
	divisionFrom := args[1]
	deponentTo := args[2]
	accountTo := args[3]
	divisionTo := args[4]
	security := args[5]
	quantity := args[6] //quantity, _ := strconv.Atoi(args[2])
	reference := args[7]
	instructionDate := args[8]
	tradeDate := args[9] // := time.Now().UTC().Unix()
	//reason := args[10]

	//accountFrom-divisionFrom-accountTo-divisionTo-security-quantity-reference-instructionDate-tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{accountFrom, divisionFrom,
		accountTo, divisionTo, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{DeponentFrom: deponentFrom, DeponentTo: deponentTo,
		Status: "initiated", Initiator: "transferer"})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(key));
}

func (t *InstructionChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting accountFrom, divisionFrom, accountTo, divisionTo, security, quantity, reference, " +
			"instructionDate, tradeDate, status")
	}

	status := args[9]

	//accountFrom-divisionFrom-accountTo-divisionTo-security-quantity-reference-instructionDate-tradeDate
	key, err := stub.CreateCompositeKey(indexName, args[0:len(args)-1])
	if err != nil {
		return shim.Error(err.Error())
	}
	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}

	var value InstructionValue
	err = json.Unmarshal(bytes, &value)
	if err != nil {
		return shim.Error(err.Error())
	}

	value.Status = status

	newBytes, err := json.Marshal(value)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, newBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil);
}

func (t *InstructionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	instructions := []Instruction{}
	for it.HasNext() {
		responseRange, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		//depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		var value InstructionValue
		err = json.Unmarshal(responseRange.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction := Instruction {
			Transferer: Source {
				Account: compositeKeyParts[0],
				Division: compositeKeyParts[1]},
			Receiver: Source {
				Account: compositeKeyParts[2],
				Division: compositeKeyParts[3]},
			Security: compositeKeyParts[4],
			Quantity: compositeKeyParts[5],
			Reference: compositeKeyParts[6],
			InstructionDate: compositeKeyParts[7],
			TradeDate: compositeKeyParts[8],
			DeponentFrom: value.DeponentFrom,
			DeponentTo: value.DeponentTo,
			Status: value.Status,
			Initiator: value.Initiator,
			Reason: Reason{Document: "", Description: "", DocumentDate: ""},
		}

		instructions = append(instructions, instruction)
	}

	result, err := json.Marshal(instructions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *InstructionChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting accountFrom, divisionFrom, accountTo, divisionTo, security, quantity, reference, " +
			"instructionDate, tradeDate")
	}

	//accountFrom-divisionFrom-accountTo-divisionTo-security-quantity-reference-instructionDate-tradeDate
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
	err := shim.Start(new(InstructionChaincode))
	if err != nil {
		logger.Errorf("Error starting Instruction chaincode: %s", err)
	}
}

