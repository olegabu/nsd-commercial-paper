package main


import (
	"fmt"
	"strconv"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"time"
	"io/ioutil"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("InstructionChaincode")

const indexName = `Instruction`

const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"

	InstructionInitiated 	= "initiated"
	InstructionMatched 		= "matched"
	InstructionExecuted 	= "executed"
	InstructionDeclined 	= "declined"
	InstructionCanceled 	= "canceled"
)

type InstructionChaincode struct {
}

// Instruction is the main data type stored in ledger
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

type Balance struct {
	Account 		string 	`json:"account"`
	Division 		string 	`json:"division"`
}

type Reason struct {
	Document 		string 	`json:"document"`
	Description 	string 	`json:"description"`
	DocumentDate 	string 	`json:"documentDate"`
}

// required for history
type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     InstructionValue  `json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

// required for history
type InstructionValue struct {
	DeponentFrom 	string 	`json:"deponentFrom"`
	DeponentTo		string 	`json:"deponentTo"`
	Status      	string 	`json:"status"`
	Initiator 		string 	`json:"initiator"`
}

// **** Instruction Methods **** //
func (this *Instruction) toStringArray() ([]string) {
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

func (this *Instruction) toCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(indexName, this.toStringArray())
}

func (this *Instruction) existsIn(stub shim.ChaincodeStubInterface) (bool) {
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

func (this *Instruction) loadFrom(stub shim.ChaincodeStubInterface) (error) {
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

func (this *Instruction) upsertIn(stub shim.ChaincodeStubInterface) (error) {
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

func (this *Instruction) setEvent(stub shim.ChaincodeStubInterface) (error) {
	bytes, err := this.toJSON()
	if err != nil {
		return err
	}

	err = stub.SetEvent(indexName + "." + this.Status, bytes)
	if err != nil {
		return err
	}

	return nil
}

func (this *Instruction) matchIf(stub shim.ChaincodeStubInterface, desiredInitiator string) pb.Response {
	err := this.loadFrom(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if this.Initiator != desiredInitiator {
		return shim.Error("Instruction is already created by " + this.Initiator)
	}

	if this.Status != InstructionInitiated {
		return shim.Error("Instruction status is not " + InstructionInitiated)
	}

	this.Status = InstructionMatched

	err = this.upsertIn(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = this.setEvent(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Instruction was successfully matched."))
}

func (this *Instruction) initiateIn(stub shim.ChaincodeStubInterface) pb.Response {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	this.Status = InstructionInitiated
	bytes, err := this.toLedgerValue()
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Instruction was successfully initiated."));
}

func (this *Instruction) fillFromCompositeKeyParts(compositeKeyParts []string) (error) {
	if len(compositeKeyParts) != 9 {
		return errors.New("Composite key parts array length must be 9.")
	}

	this.Transferer.Account 	= compositeKeyParts[0]
	this.Transferer.Division 	= compositeKeyParts[1]
	this.Receiver.Account 		= compositeKeyParts[2]
	this.Receiver.Division 		= compositeKeyParts[3]
	this.Security 				= compositeKeyParts[4]
	this.Quantity 				= compositeKeyParts[5]
	this.Reference 				= compositeKeyParts[6]
	this.InstructionDate 		= compositeKeyParts[7]
	this.TradeDate 				= compositeKeyParts[8]

	return nil
}

func (this *Instruction) fillFromArgs(args []string) (error) {
	if len(args) < 11 {
		return errors.New("Incorrect number of arguments. Expecting >= 11.")
	}

	this.DeponentFrom			= args[0]
	this.Transferer.Account 	= args[1]
	this.Transferer.Division 	= args[2]
	this.DeponentTo				= args[3]
	this.Receiver.Account 		= args[4]
	this.Receiver.Division 		= args[5]
	this.Security 				= args[6]
	this.Quantity 				= args[7]
	this.Reference 				= args[8]
	this.InstructionDate 		= args[9]
	this.TradeDate 				= args[10]
	//this.Reason					= args[11]

	return nil
}

func (this *Instruction) toLedgerValue() ([]byte, error) {
	return json.Marshal([]string{this.DeponentFrom, this.DeponentTo, this.Status, this.Initiator})
}

func (this *Instruction) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Instruction) fillFromLedgerValue(bytes []byte) (error) {
	var str []string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	this.DeponentFrom = str[0]
	this.DeponentTo = str[1]
	this.Status = str[2]
	this.Initiator = str[3]

	return nil
}

// **** Chaincode Methods **** //
func (t *InstructionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### InstructionChaincode Init ###########")

	args := stub.GetArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting bookChannel")
	}

	stub.PutState ("bookChannel", args[1])

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

func (t *InstructionChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	if instruction.existsIn(stub) {
		return instruction.matchIf(stub, InitiatorIsTransferer)
	} else {
		instruction.Initiator = InitiatorIsReceiver
		instruction.Status = InstructionInitiated
		if instruction.upsertIn(stub) != nil {
			return shim.Error("Instruction initialization error.")
		}
		return shim.Success([]byte("Instruction created."))
	}
}

func (t *InstructionChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	if instruction.existsIn(stub) {
		return instruction.matchIf(stub, InitiatorIsReceiver)
	} else {
		instruction.Initiator = InitiatorIsTransferer
		instruction.Status = InstructionInitiated
		if instruction.upsertIn(stub) != nil {
			return shim.Error("Instruction initialization error.")
		}
		return shim.Success([]byte("Instruction created."))
	}
}

func (t *InstructionChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	status := args[len(args) - 1]

	if status != InstructionExecuted {
		shim.Error("Wrong instruction state.")
	}

	if instruction.existsIn(stub) {
		err := instruction.loadFrom(stub)
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction.Status = status

		err = instruction.upsertIn(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		return shim.Error("Instruction state change error.")
	}
	return shim.Success(nil)
}

func (t *InstructionChaincode) check(stub shim.ChaincodeStubInterface, account string, division string, security string,
	quantity int) bool {

	myOrganization := getMyOrganization()
	logger.Debugf("ORGANIZATION IS: " + myOrganization)

	if  myOrganization == "nsd.nsd.ru" {
		/*bookChannelBytes, err := stub.GetState("bookChannel")
		if err != nil {
			logger.Error("cannot find bookChannel")
			return false
		}*/

		byteArgs := [][]byte{}
		byteArgs = append(byteArgs, []byte("check"))
		byteArgs = append(byteArgs, []byte(account))
		byteArgs = append(byteArgs, []byte(division))
		byteArgs = append(byteArgs, []byte(security))
		byteArgs = append(byteArgs, []byte(strconv.Itoa(quantity)))

		logger.Debugf("BEFORE INVOKE")

		res := stub.InvokeChaincode("book", byteArgs, /*string(bookChannelBytes)*/"depository")
		if res.GetStatus() != 200 {
			return false
		}

		return true
	}

	return true
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

		instruction := Instruction{}

		err = instruction.fillFromLedgerValue(response.Value)
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = instruction.fillFromCompositeKeyParts(compositeKeyParts)
		if err != nil {
			return shim.Error(err.Error())
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

	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := instruction.toCompositeKey(stub)
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

// **** Security Methods **** //
func getOrganization(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----"):strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]

	return organization
}

func getCreatorOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	return getOrganization(certificate)
}

func getMyOrganization() string {
	// TODO get the filename from $CORE_PEER_TLS_ROOTCERT_FILE
	// better way perhaps is to pass a flag in transient map to nsd peer to ask to check against book chaincode
	certFilename := "/etc/hyperledger/fabric/peer.crt"
	certificate, err := ioutil.ReadFile(certFilename)
	if err != nil {
		logger.Debugf("cannot read my peer's certificate file %s", certFilename)
		return ""
	}

	return getOrganization(certificate)
}

// **** main method **** //
func main() {
	err := shim.Start(new(InstructionChaincode))
	if err != nil {
		logger.Errorf("Error starting Instruction chaincode: %s", err)
	}
}

