package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"
	"time"

	"bytes"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("InstructionChaincode")

const instructionIndex = `Instruction`
const authenticationIndex = `Authentication`

const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"

	InstructionInitiated = "initiated"
	InstructionMatched   = "matched"
	InstructionSigned    = "signed"
	InstructionExecuted  = "executed"
	InstructionDeclined  = "declined"
	InstructionCanceled  = "canceled"
)

type InstructionChaincode struct {
}

// Instruction is the main data type stored in ledger
type Instruction struct {
	Key   InstructionKey   `json:"key"`
	Value InstructionValue `json:"value"`
}

type InstructionKey struct {
	Transferer Balance `json:"transferer"`
	Receiver   Balance `json:"receiver"`
	Security   string  `json:"security"`
	//TODO should be int like everywhere
	Quantity        string `json:"quantity"`
	Reference       string `json:"reference"`
	InstructionDate string `json:"instructionDate"`
	TradeDate       string `json:"tradeDate"`
}

type InstructionValue struct {
	DeponentFrom         string `json:"deponentFrom"`
	DeponentTo           string `json:"deponentTo"`
	Status               string `json:"status"`
	Initiator            string `json:"initiator"`
	ReasonFrom           Reason `json:"reasonFrom"`
	ReasonTo             Reason `json:"reasonTo"`
	AlamedaFrom          string `json:"alamedaFrom"`
	AlamedaTo            string `json:"alamedaTo"`
	AlamedaSignatureFrom string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo   string `json:"alamedaSignatureTo"`
}

type Balance struct {
	Account  string `json:"account"`
	Division string `json:"division"`
}

type Reason struct {
	DocumentDate string `json:"created"`
	Document     string `json:"document"`
	Description  string `json:"description"`
}

// required for history
type KeyModificationValue struct {
	TxId      string           `json:"txId"`
	Value     InstructionValue `json:"value"`
	Timestamp string           `json:"timestamp"`
	IsDelete  bool             `json:"isDelete"`
}

// **** Instruction Methods **** //
func (this *Instruction) toCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	keyParts := []string{
		this.Key.Transferer.Account,
		this.Key.Transferer.Division,
		this.Key.Receiver.Account,
		this.Key.Receiver.Division,
		this.Key.Security,
		this.Key.Quantity,
		this.Key.Reference,
		this.Key.InstructionDate,
		this.Key.TradeDate,
	}
	return stub.CreateCompositeKey(instructionIndex, keyParts)
}

func (this *Instruction) fillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < 9 {
		return errors.New("Composite key parts array length must be at least 9.")
	}

	this.Key.Transferer.Account = compositeKeyParts[0]
	this.Key.Transferer.Division = compositeKeyParts[1]
	this.Key.Receiver.Account = compositeKeyParts[2]
	this.Key.Receiver.Division = compositeKeyParts[3]
	this.Key.Security = compositeKeyParts[4]
	this.Key.Quantity = compositeKeyParts[5]
	this.Key.Reference = compositeKeyParts[6]
	this.Key.InstructionDate = compositeKeyParts[7]
	this.Key.TradeDate = compositeKeyParts[8]

	return nil
}

func (this *Instruction) fillFromArgs(args []string) error {
	return this.fillFromCompositeKeyParts(args)
}

func (this *Instruction) existsIn(stub shim.ChaincodeStubInterface) bool {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return false
	}

	if _, err := stub.GetState(compositeKey); err != nil {
		return false
	}

	return true
}

func (this *Instruction) loadFrom(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return err
	}

	data, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return this.fillFromLedgerValue(data)
}

func (this *Instruction) upsertIn(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return err
	}

	value, err := this.toLedgerValue()
	if err != nil {
		return err
	}

	if err = stub.PutState(compositeKey, value); err != nil {
		return err
	}

	return nil
}

func (this *Instruction) setEvent(stub shim.ChaincodeStubInterface) error {
	data, err := this.toJSON()
	if err != nil {
		return err
	}

	if err = stub.SetEvent(instructionIndex+"."+this.Value.Status, data); err != nil {
		return err
	}

	return nil
}

func (this *Instruction) matchIf(stub shim.ChaincodeStubInterface, desiredInitiator string) pb.Response {
	if err := this.loadFrom(stub); err != nil {
		return pb.Response{Status: 404, Message: "Instruction not found."}
	}

	if this.Value.Initiator != desiredInitiator {
		return pb.Response{Status: 400, Message: "Instruction is already created by " + this.Value.Initiator}
	}

	if this.Value.Status != InstructionInitiated {
		return pb.Response{Status: 400, Message: "Instruction status is not " + InstructionInitiated}
	}

	this.Value.Status = InstructionMatched

	this.Value.AlamedaFrom, this.Value.AlamedaTo = this.createAlamedaXMLs()

	if err := this.upsertIn(stub); err != nil {
		return pb.Response{Status: 520, Message: "Persistence failure."}
	}

	if err := this.setEvent(stub); err != nil {
		return pb.Response{Status: 520, Message: "Event emission failure."}
	}

	return shim.Success(nil)
}

func (this *Instruction) initiateIn(stub shim.ChaincodeStubInterface) pb.Response {
	compositeKey, err := this.toCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	this.Value.Status = InstructionInitiated
	data, err := this.toLedgerValue()
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, data)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Instruction was successfully initiated."))
}

func (this *Instruction) toLedgerValue() ([]byte, error) {
	return json.Marshal(this.Value)
}

func (this *Instruction) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Instruction) fillFromLedgerValue(bytes []byte) error {
	if err := json.Unmarshal(bytes, &this.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Instruction) createAlamedaXMLs() (string, string) {
	const xmlTemplate = `
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>{{.Depositary}}</deposit_c>
<contrag_c>{{.Initiator}}</contrag_c>
<contr_d_id>{{.InstructionID}}</contr_d_id>
<createdate>{{.Instruction.InstructionDate}}</createdate>
<order_t_id>{{.OperationCode}}</order_t_id>
<execute_dt>{{.Instruction.InstructionDate}}</execute_dt>
<expirat_dt>{{.ExpirationDate}}</expirat_dt>
</ORDER_HEADER>
<MF010>
<dep_acc_c>{{.Instruction.Transferer.Account}}</dep_acc_c>
<sec_c>{{.Instruction.Transferer.Division}}</sec_c>
<deponent_c>{{.Instruction.DeponentFrom}}</deponent_c>
<corr_acc_c>{{.Instruction.Receiver.Account}}</corr_acc_c>
<corr_sec_c>{{.Instruction.Receiver.Division}}</corr_sec_c>
<corr_code>{{.Instruction.DeponentTo}}</corr_code>
<based_on>{{.Instruction.Reason.Description}}</based_on>
<based_numb>{{.Instruction.Reason.Document}}</based_numb>
<based_date>{{.Instruction.Reason.DocumentDate}}</based_date>
<securities><security>
<security_c>{{.Instruction.Security}}</security_c>
<security_q>{{.Instruction.Quantity}}</security_q>
</security>
</securities>
<deal_reference>{{.Instruction.Reference}}</deal_reference>
<date_deal>{{.Instruction.TradeDate}}</date_deal>
</MF010>
`
	type InstructionWrapper struct {
		Instruction    Instruction
		Depositary     string
		Initiator      string
		InstructionID  string //TODO: remove this
		OperationCode  string
		ExpirationDate string
	}

	instructionWrapper := InstructionWrapper{
		Instruction:    *this,
		Depositary:     "DEPOSITARY_CODE",
		Initiator:      this.Value.DeponentFrom,
		InstructionID:  "42",
		OperationCode:  "16",
		ExpirationDate: "+24h",
	}

	t := template.Must(template.New("xmlTemplate").Parse(xmlTemplate))

	buf := new(bytes.Buffer)
	t.Execute(buf, instructionWrapper)
	alamedaFrom := buf.String()

	buf.Reset()
	instructionWrapper.OperationCode = "16/1"
	instructionWrapper.Initiator = this.Value.DeponentTo

	t.Execute(buf, instructionWrapper)
	alamedaTo := buf.String()

	return alamedaFrom, alamedaTo
}

// **** Chaincode Methods **** //
func (t *InstructionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### InstructionChaincode Init ###########")

	args := stub.GetStringArgs()
	logger.Info("########### " + strings.Join(args, " ") + " ###########")
	logger.Info("########### " + getCreatorOrganization(stub) + " ###########")

	type Organization struct {
		Name     string    `json:"organization"`
		Balances []Balance `json:"balances"`
	}

	var organizations []Organization
	if err := json.Unmarshal([]byte(args[1]), &organizations); err == nil && len(organizations) != 0 {
		for _, organization := range organizations {
			for _, balance := range organization.Balances {
				keyParts := []string{balance.Account, balance.Division}
				if key, err := stub.CreateCompositeKey(authenticationIndex, keyParts); err == nil {
					stub.PutState(key, []byte(organization.Name))
				}
			}
		}
	}

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
	if function == "sign" {
		return t.sign(stub, args)
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

	err := fmt.Sprintf("Unknown function, check the first argument, must be one of: "+
		"receive, transfer, query, history, status, sign. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func (t *InstructionChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	if err := instruction.fillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	if authenticateCaller(stub, instruction.Key.Receiver) == false {
		return pb.Response{Status: 403, Message: "Caller must be receiver."}
	}

	if instruction.existsIn(stub) {
		return instruction.matchIf(stub, InitiatorIsTransferer)
	} else {
		instruction.Value.DeponentFrom = args[9]
		instruction.Value.DeponentTo = args[10]
		instruction.Value.Initiator = InitiatorIsReceiver
		instruction.Value.Status = InstructionInitiated
		if instruction.upsertIn(stub) != nil {
			return pb.Response{Status: 520, Message: "Persistence failure."}

		}
		return shim.Success(nil)
	}
}

func (t *InstructionChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	if err := instruction.fillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	if authenticateCaller(stub, instruction.Key.Transferer) == false {
		return pb.Response{Status: 403, Message: "Caller must be transferer."}
	}

	if instruction.existsIn(stub) {
		return instruction.matchIf(stub, InitiatorIsReceiver)
	} else {
		instruction.Value.DeponentFrom = args[9]
		instruction.Value.DeponentTo = args[10]
		instruction.Value.Initiator = InitiatorIsTransferer
		instruction.Value.Status = InstructionInitiated
		if instruction.upsertIn(stub) != nil {
			return pb.Response{Status: 520, Message: "Persistence failure."}

		}
		return shim.Success(nil)
	}
}

func (t *InstructionChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("########### InstructionChaincode status ###########")
	logger.Info(args)

	instruction := Instruction{}
	if err := instruction.fillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
	}

	status := args[len(args)-1]

	callerIsTransferer := authenticateCaller(stub, instruction.Key.Transferer)
	callerIsReceiver := authenticateCaller(stub, instruction.Key.Receiver)
	callerIsNSD := getCreatorOrganization(stub) == "nsd.nsd.ru"

	if callerIsTransferer {
		logger.Info("callerIsTransferer")
	}
	if callerIsReceiver {
		logger.Info("callerIsReceiver")
	}
	if callerIsNSD {
		logger.Info("callerIsNSD")
	}

	if !(callerIsTransferer || callerIsReceiver || callerIsNSD) {
		return pb.Response{Status: 403, Message: "Instruction status can be changed either by transferer, receiver or NSD."}
	}

	if err := instruction.loadFrom(stub); err != nil {
		return pb.Response{Status: 404, Message: "Instruction not found."}
	}

	switch {
	case callerIsNSD && status == InstructionDeclined, callerIsNSD && status == InstructionExecuted:
		instruction.Value.Status = status
		if err := instruction.upsertIn(stub); err != nil {
			return pb.Response{Status: 520, Message: "Persistence failure."}
		}
	case (callerIsTransferer || callerIsReceiver) && instruction.Value.Status == InstructionInitiated && status == InstructionCanceled:
		if (callerIsTransferer && instruction.Value.Initiator == InitiatorIsTransferer) || (callerIsReceiver && instruction.Value.Initiator == InitiatorIsReceiver) {
			instruction.Value.Status = status
			if err := instruction.upsertIn(stub); err != nil {
				return pb.Response{Status: 520, Message: "Persistence failure."}
			}
		}
	default:
		return pb.Response{Status: 406, Message: "Instruction status or caller identity is wrong."}
	}

	return shim.Success(nil)
}

func (t *InstructionChaincode) check(stub shim.ChaincodeStubInterface, account string, division string, security string,
	quantity int) bool {

	myOrganization := getMyOrganization()
	logger.Debugf("ORGANIZATION IS: " + myOrganization)

	if myOrganization == "nsd.nsd.ru" {
		byteArgs := [][]byte{}
		byteArgs = append(byteArgs, []byte("check"))
		byteArgs = append(byteArgs, []byte(account))
		byteArgs = append(byteArgs, []byte(division))
		byteArgs = append(byteArgs, []byte(security))
		byteArgs = append(byteArgs, []byte(strconv.Itoa(quantity)))

		logger.Debugf("BEFORE INVOKE")

		res := stub.InvokeChaincode("book", byteArgs, "depository")
		if res.GetStatus() != 200 {
			return false
		}

		return true
	}

	return true
}

func (t *InstructionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(instructionIndex, []string{})
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

		if err := instruction.fillFromLedgerValue(response.Value); err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		if err := instruction.fillFromCompositeKeyParts(compositeKeyParts); err != nil {
			return shim.Error(err.Error())
		}

		callerIsTransferer := authenticateCaller(stub, instruction.Key.Transferer)
		callerIsReceiver := authenticateCaller(stub, instruction.Key.Receiver)
		if !(callerIsTransferer || callerIsReceiver) {
			continue
		}

		if (callerIsTransferer && instruction.Value.Initiator == InitiatorIsTransferer) ||
			(callerIsReceiver && instruction.Value.Initiator == InitiatorIsReceiver) ||
			(instruction.Value.Status == InstructionMatched) ||
			(instruction.Value.Status == InstructionSigned) ||
			(instruction.Value.Status == InstructionDeclined) {
			instructions = append(instructions, instruction)
		}
	}

	result, err := json.Marshal(instructions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *InstructionChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	if err := instruction.fillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
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

func (t *InstructionChaincode) sign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := Instruction{}
	if err := instruction.fillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
	}

	signature := args[len(args)-1]

	if err := instruction.loadFrom(stub); err != nil {
		return pb.Response{Status: 404, Message: "Instruction not found."}
	}

	callerIsTransferer := authenticateCaller(stub, instruction.Key.Transferer)
	callerIsReceiver := authenticateCaller(stub, instruction.Key.Receiver)

	if !(callerIsTransferer || callerIsReceiver) {
		return pb.Response{Status: 403}
	}

	if callerIsTransferer {
		instruction.Value.AlamedaSignatureFrom = signature
	}

	if callerIsReceiver {
		instruction.Value.AlamedaSignatureTo = signature
	}

	if instruction.Value.AlamedaSignatureFrom != "" && instruction.Value.AlamedaSignatureTo != "" {
		instruction.Value.Status = InstructionSigned
		instruction.setEvent(stub)
	}

	if err := instruction.upsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// **** Security Methods **** //
func getOrganization(certificate []byte) string {
	logger.Info("########### InstructionChaincode getOrganization ###########")
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	logger.Info("getOrganization: " + organization)

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

func authenticateCaller(stub shim.ChaincodeStubInterface, callerBalance Balance) bool {
	keyParts := []string{callerBalance.Account, callerBalance.Division}
	if key, err := stub.CreateCompositeKey(authenticationIndex, keyParts); err == nil {
		if data, err := stub.GetState(key); err == nil && getCreatorOrganization(stub) == string(data) {
			return true
		}
	}
	return false
}

// **** main method **** //
func main() {
	err := shim.Start(new(InstructionChaincode))
	if err != nil {
		logger.Errorf("Error starting Instruction chaincode: %s", err)
	}
}
