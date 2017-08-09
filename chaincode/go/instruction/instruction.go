package main


import (
	"fmt"
	"strconv"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("InstructionChaincode")

const indexName = `Instruction`

// InstructionChaincode
type InstructionChaincode struct {
}

type InstructionKey struct {
	Transferer 		Balance
	Receiver		Balance
	Security 		string
	Quantity 		string
	Reference 		string
	InstructionDate string
	TradeDate 		string
}

func (this *InstructionKey) ToStringArray() ([]string) {
	return []string{
		this.Transferer.Account,
		this.Transferer.Division,
		this.Receiver.Account,
		this.Receiver.Division,
		this.Security,
		this.Quantity,
		this.Reference,
		this.InstructionDate,
		this.TradeDate,
	}
}

func (this *InstructionKey) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(indexName, this.ToStringArray())
}

type InstructionValue struct {
	DeponentFrom 	string 	`json:"deponentFrom"`
	DeponentTo		string 	`json:"deponentTo"`
	Status      	string 	`json:"status"`
	Initiator 		string 	`json:"initiator"`
}

const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"
)

const (
	ArgIndexTransfererDeponent 	= 0
	ArgIndexTransfererAccount 	= 1
	ArgIndexTransfererDivision 	= 2
	ArgIndexReceiverDeponent 	= 3
	ArgIndexReceiverAccount 	= 4
	ArgIndexReceiverDivision 	= 5
	ArgIndexSecurity 			= 6
	ArgIndexQuantity 			= 7
	ArgIndexReference 			= 8
	ArgIndexInstructionDate 	= 9
	ArgIndexTradeDate 			= 10
	ArgIndexReason 				= 11
)

const (
	InstructionInitiated 	= "initiated"
	InstructionMatched 		= "matched"
	InstructionExecuted 	= "executed"
	InstructionDeclined 	= "declined"
	InstructionCanceled 	= "canceled"
)

type Balance struct {
	Account 		string 	`json:"account"`
	Division 		string 	`json:"division"`
}

type Reason struct {
	Document 		string 	`json:"document"`
	Description 	string 	`json:"description"`
	DocumentDate 	string 	`json:"documentDate"`
}

type Instruction struct {
	Transferer      Balance `json:"transferer"`
	Receiver        Balance `json:"receiver"`
	Security        string 	`json:"security"`
	Quantity        string 	`json:"quantity"`
	Reference       string 	`json:"reference"`
	InstructionDate string 	`json:"instructionDate"`
	TradeDate       string 	`json:"tradeDate"`
	DeponentFrom    string  `json:"deponentFrom"`
	DeponentTo      string  `json:"deponentTo"`
	Status          string 	`json:"status"`
	Initiator       string 	`json:"initiator"`
	Reason          Reason 	`json:"reason"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     InstructionValue  `json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

func (t *InstructionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### InstructionChaincode Init ###########")
/*
	args := stub.GetArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting bookChannel")
	}

	stub.PutState ("bookChannel", args[1])
*/
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

	// for debugging only
	if function == "check" {
		if len(args) != 4 {
			return shim.Error("Incorrect number of arguments. " +
				"Expecting account, division, security, quantity")
		}
		quantity, _ := strconv.Atoi(args[3])

		if t.check(stub, args[0], args[1], args[2], quantity) {
			return shim.Success(nil)
		} else {
			return shim.Error("book returned false")
		}
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
	// deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12.")
	}

	key := InstructionKey{
		Transferer: Balance{
			Account:		args[ArgIndexTransfererAccount],
			Division: 		args[ArgIndexTransfererDivision]},
		Receiver: Balance{
			Account:		args[ArgIndexReceiverAccount],
			Division: 		args[ArgIndexReceiverDivision]},
		Security: 			args[ArgIndexSecurity],
		Quantity: 			args[ArgIndexQuantity],
		Reference: 			args[ArgIndexReference],
		InstructionDate: 	args[ArgIndexInstructionDate],
		TradeDate: 			args[ArgIndexTradeDate],
	}

	compositeKey, err := key.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, _ := stub.GetState(compositeKey)
	if bytes == nil {
		// Instruction does not exist
		return t.initiate(stub, key, InstructionValue{
			DeponentFrom: 	args[ArgIndexTransfererDeponent],
			DeponentTo: 	args[ArgIndexReceiverDeponent],
			Status:			InstructionInitiated,
			Initiator:		InitiatorIsReceiver})
	} else {
		// Instruction do exist - match it
		var value InstructionValue
		err = json.Unmarshal(bytes, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		if value.Initiator == InitiatorIsTransferer {
			value.Status = InstructionMatched
			return t.match(stub, key, value)
		} else {
			return shim.Error("Access denied.")
		}
	}
}

func (t *InstructionChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12.")
	}

	key := InstructionKey{
		Transferer: Balance{
			Account:		args[ArgIndexTransfererAccount],
			Division: 		args[ArgIndexTransfererDivision]},
		Receiver: Balance{
			Account:		args[ArgIndexReceiverAccount],
			Division: 		args[ArgIndexReceiverDivision]},
		Security: 			args[ArgIndexSecurity],
		Quantity: 			args[ArgIndexQuantity],
		Reference: 			args[ArgIndexReference],
		InstructionDate: 	args[ArgIndexInstructionDate],
		TradeDate: 			args[ArgIndexTradeDate],
	}

	compositeKey, err := key.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, _ := stub.GetState(compositeKey)
	if bytes == nil {
		// Instruction does not exist
		return t.initiate(stub, key, InstructionValue{
			DeponentFrom: 	args[ArgIndexTransfererDeponent],
			DeponentTo: 	args[ArgIndexReceiverDeponent],
			Status:			InstructionInitiated,
			Initiator:		InitiatorIsTransferer})
	} else {
		// Instruction do exist - match it
		var value InstructionValue
		err = json.Unmarshal(bytes, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		if value.Initiator == InitiatorIsReceiver {
			value.Status = InstructionMatched
			return t.match(stub, key, value)
		} else {
			return shim.Error("Access denied.")
		}
	}
}

func (t *InstructionChaincode) initiate(stub shim.ChaincodeStubInterface, key InstructionKey, value InstructionValue) pb.Response {
	compositeKey, err := key.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	value.Status = InstructionInitiated
	bytes, err := json.Marshal(value)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Instruction was successfully initiated."));
}

func (t *InstructionChaincode) match(stub shim.ChaincodeStubInterface, key InstructionKey, value InstructionValue) pb.Response {
	compositeKey, err := key.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	value.Status = InstructionMatched
	bytes, err := json.Marshal(value)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Instruction was successfully matched."));
}

func (t *InstructionChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0.")
	}

	return shim.Success([]byte("METHOD IS NOT READY."));
}

func (t *InstructionChaincode) check(stub shim.ChaincodeStubInterface, account string, division string, security string,
	quantity int) bool {

	bookChannelBytes, err := stub.GetState("bookChannel")
	if err != nil {
		logger.Error("cannot find bookChannel")
		return false
	}

	byteArgs := [][]byte{}
	byteArgs = append(byteArgs, []byte("check"))
	byteArgs = append(byteArgs, []byte(account))
	byteArgs = append(byteArgs, []byte(division))
	byteArgs = append(byteArgs, []byte(security))
	byteArgs = append(byteArgs, []byte(strconv.Itoa(quantity)))

	res := stub.InvokeChaincode("book", byteArgs, string(bookChannelBytes))

	logger.Debug(res)

	return res.Status == 200
}

func (t *InstructionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	instructions := []Instruction{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		var value InstructionValue
		err = json.Unmarshal(response.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction := Instruction{
			Transferer: Balance{
				Account: compositeKeyParts[0],
				Division: compositeKeyParts[1]},
			Receiver: Balance{
				Account: compositeKeyParts[2],
				Division: compositeKeyParts[3]},
			Security: compositeKeyParts[4],
			Quantity: compositeKeyParts[5],
			Reference: compositeKeyParts[6],
			InstructionDate: compositeKeyParts[7],
			TradeDate: compositeKeyParts[8],
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
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo, " +
			"security, quantity, reference, instructionDate, tradeDate")
	}

	key := InstructionKey{
		Transferer: Balance{
			Account:		args[ArgIndexTransfererAccount],
			Division: 		args[ArgIndexTransfererDivision]},
		Receiver: Balance{
			Account:		args[ArgIndexReceiverAccount],
			Division: 		args[ArgIndexReceiverDivision]},
		Security: 			args[ArgIndexSecurity],
		Quantity: 			args[ArgIndexQuantity],
		Reference: 			args[ArgIndexReference],
		InstructionDate: 	args[ArgIndexInstructionDate],
		TradeDate: 			args[ArgIndexTradeDate],
	}

	compositeKey, err := key.ToCompositeKey(stub)
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

func main() {
	err := shim.Start(new(InstructionChaincode))
	if err != nil {
		logger.Errorf("Error starting Instruction chaincode: %s", err)
	}
}

