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
const indexName = `depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate`

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type InstructionValue struct {
	Status      		string 	`json:"status"`
	Initiator 			string 	`json:"initiator"`
}

type UserId struct {
	Deponent 	string 	`json:"dep"`
	Account 	string 	`json:"acc"`
	Division 	string 	`json:"div"`
}

type Reason struct {
	Document 	string 	`json:"document"`
	Description string 	`json:"description"`
	Created 	string 	`json:"created"`
}

type Instruction struct {
	Transferer 		UserId 	`json:"transferer"`
	Receiver   		UserId 	`json:"receiver"`
	Security   		string 	`json:"security"`
	Quantity   		string 	`json:"quantity"`
	Reference  		string 	`json:"reference"`
	InstructionDate string 	`json:"instruction_date"`
	TradeDate  		string 	`json:"trade_date"`
	Status     		string 	`json:"status"`
	Initiator 		string 	`json:"initiator"`
	Reason 			Reason 	`json:"reason"`
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

func (t *SimpleChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// deponentFrom, accountFrom, divisionFrom, *, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11.")
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

	//depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate
	key, err := stub.CreateCompositeKey(indexName, []string{deponentFrom, accountFrom, divisionFrom,
		deponentTo, accountTo, divisionTo, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{Status: "initiated", Initiator: "receiver"})
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
	// *, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo,
	// security, quantity, reference, instructionDate, tradeDate, reason
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11.")
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

	//depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate
	key, err := stub.CreateCompositeKey(indexName, []string{deponentFrom, accountFrom, divisionFrom,
		deponentTo, accountTo, divisionTo, security, quantity, reference, instructionDate, tradeDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(InstructionValue{Status: "initiated", Initiator: "transferer"})
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
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12.")
	}

	status := args[11]

	//depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate
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

		instruction := Instruction{
			Transferer: UserId{
				Deponent: compositeKeyParts[0],
				Account: compositeKeyParts[1],
				Division: compositeKeyParts[2]},
			Receiver: UserId{
				Deponent: compositeKeyParts[3],
				Account: compositeKeyParts[4],
				Division: compositeKeyParts[5]},
			Security: compositeKeyParts[6],
			Quantity: compositeKeyParts[7],
			Reference: compositeKeyParts[8],
			InstructionDate: compositeKeyParts[9],
			TradeDate: compositeKeyParts[10],
			Status: value.Status,
			Initiator: value.Initiator,
			Reason: Reason{Document: "", Description: "", Created: ""},
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
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11.")
	}

	//depF-accF-divF-depT-accT-divT-sec-qua-ref-insDate-traDate
	key, err := stub.CreateCompositeKey(indexName, args)
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

