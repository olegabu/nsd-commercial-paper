package main


import (
	"fmt"
	//"strconv"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("instruction_sc")
const indexName = `from-to-security-quantity-reference-instructionDate-tradeDate`

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type InstructionValue struct {
	Status      string 	`json:"status"`
	InitiatorID string 	`json:"initiatorID"`
}

type Instruction struct {
	Transferer 	string 	`json:"transferer"`
	Receiver   	string 	`json:"receiver"`
	Security   	string 	`json:"security"`
	Quantity   	string 	`json:"quantity"`
	Reference  	string 	`json:"reference"`
	InstructionDate string 	`json:"instruction_date"`
	TradeDate  	string 	`json:"trade_date"`
	Status     	string 	`json:"status"`
	InitiatorID string 	`json:"initiatorID"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"tx_id"`
	Value     InstructionValue  `json:"value"`
	Timestamp string 			`json:"timestamp`
	IsDelete  bool   			`json:"is_delete`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### Instruction Smart Contract Init ###########")

	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Instruction Smart Contract Invoke ###########")

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

	logger.Errorf("Unknown action, check the first argument, must be one of 'receive', 'transfer', 'query','history', or " +
		"'status'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'from', 'to', 'query'," +
		"'history', or 'status'. But got: %v", args[0]))
}

func (t *SimpleChaincode) GetOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	data := certificate[strings.Index(string(certificate), "-----"):strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]

	return organization
}

// receive(from, security, quantity, reference, instructionDate, tradeDate, reason) returns instruction id
// to is taken from transaction creator’s cert
func (t *SimpleChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7.")
	}

	to := t.GetOrganization(stub)
	from := args[0]
	security := args[1]
	quantity := args[2] //quantity, _ := strconv.Atoi(args[2])
	reference := args[3]
	instructionDate := args[4] // := time.Now().UTC().Unix()
	tradeDate := args[5]
	//reason := args[6]

	//from-to-security-quantity-reference-instructionDate-tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{from, to, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{Status: "initiated", InitiatorID: to})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(key));
}

func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7.")
	}

	from := t.GetOrganization(stub)
	to := args[0]
	security := args[1]
	quantity := args[2] //quantity, _ := strconv.Atoi(args[2])
	reference := args[3]
	instructionDate := args[4] // := time.Now().UTC().Unix()
	tradeDate := args[5]
	//reason := args[6]

	//from-to-security-quantity-reference-instructionDate-tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{from, to, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{Status: "initiated", InitiatorID: from})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(key));
}

func (t *SimpleChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8.")
	}

	status := args[7]

	//from, to, security, quantity, reference, instructionDate, tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{args[0], args[1], args[2], args[3], args[4], args[5], args[6]})
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

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		var value InstructionValue
		err = json.Unmarshal(responseRange.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction := Instruction{
			Transferer: compositeKeyParts[0],
			Receiver: compositeKeyParts[1],
			Security: compositeKeyParts[2],
			Quantity: compositeKeyParts[3],
			Reference: compositeKeyParts[4],
			InstructionDate: compositeKeyParts[5],
			TradeDate: compositeKeyParts[6],
			Status: value.Status,
			InitiatorID: value.InitiatorID,
		}

		instructions = append(instructions, instruction)

	}

	result, err := json.Marshal(instructions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *SimpleChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7.")
	}

	//from, to, security, quantity, reference, instructionDate, tradeDate
	key, err := stub.CreateCompositeKey(indexName, []string{args[0], args[1], args[2], args[3], args[4], args[5], args[6]})
	if err != nil {
		return shim.Error(err.Error())
	}

	//organization := t.GetOrganization(stub)

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
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Instruction chaincode: %s", err)
	}
}

