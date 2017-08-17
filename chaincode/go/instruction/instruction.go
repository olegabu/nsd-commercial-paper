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
	Transferer Balance `json:"transferer"`
	Receiver   Balance `json:"receiver"`
	Security   string  `json:"security"`
	//TODO should be int like everywhere
	Quantity             string `json:"quantity"`
	Reference            string `json:"reference"`
	InstructionDate      string `json:"instructionDate"`
	TradeDate            string `json:"tradeDate"`
	DeponentFrom         string `json:"deponentFrom"`
	DeponentTo           string `json:"deponentTo"`
	Status               string `json:"status"`
	Initiator            string `json:"initiator"`
	Reason               Reason `json:"reason"`
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
	Document     string `json:"document"`
	Description  string `json:"description"`
	DocumentDate string `json:"documentDate"`
}

// required for history
type KeyModificationValue struct {
	TxId      string           `json:"txId"`
	Value     InstructionValue `json:"value"`
	Timestamp string           `json:"timestamp"`
	IsDelete  bool             `json:"isDelete"`
}

// required for history
type InstructionValue struct {
	DeponentFrom         string `json:"deponentFrom"`
	DeponentTo           string `json:"deponentTo"`
	Status               string `json:"status"`
	Initiator            string `json:"initiator"`
	AlamedaFrom          string `json:"alamedaFrom"`
	AlamedaTo            string `json:"alamedaTo"`
	AlamedaSignatureFrom string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo   string `json:"alamedaSignatureTo"`
}

// **** Instruction Methods **** //
func (this *Instruction) toStringArray() []string {
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
	return stub.CreateCompositeKey(instructionIndex, this.toStringArray())
}

func (this *Instruction) existsIn(stub shim.ChaincodeStubInterface) bool {
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

func (this *Instruction) loadFrom(stub shim.ChaincodeStubInterface) error {
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

func (this *Instruction) upsertIn(stub shim.ChaincodeStubInterface) error {
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

func (this *Instruction) setEvent(stub shim.ChaincodeStubInterface) error {
	bytes, err := this.toJSON()
	if err != nil {
		return err
	}

	err = stub.SetEvent(instructionIndex+"."+this.Status, bytes)
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
		return pb.Response{Status: 400, Message: "Instruction is already created by " + this.Initiator}
	}

	if this.Status != InstructionInitiated {
		return pb.Response{Status: 400, Message: "Instruction status is not " + InstructionInitiated}
	}

	this.Status = InstructionMatched

	this.AlamedaFrom, this.AlamedaTo = this.createAlamedaXMLs()

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

	return shim.Success([]byte("Instruction was successfully initiated."))
}

func (this *Instruction) fillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < 9 {
		return errors.New("Composite key parts array length must be at least 9.")
	}

	this.Transferer.Account = compositeKeyParts[0]
	this.Transferer.Division = compositeKeyParts[1]
	this.Receiver.Account = compositeKeyParts[2]
	this.Receiver.Division = compositeKeyParts[3]
	this.Security = compositeKeyParts[4]
	this.Quantity = compositeKeyParts[5]
	this.Reference = compositeKeyParts[6]
	this.InstructionDate = compositeKeyParts[7]
	this.TradeDate = compositeKeyParts[8]

	return nil
}

func (this *Instruction) fillFromArgs(args []string) error {
	if len(args) < 11 {
		return errors.New("Incorrect number of arguments. Expecting >= 11.")
	}

	this.DeponentFrom = args[0]
	this.Transferer.Account = args[1]
	this.Transferer.Division = args[2]
	this.DeponentTo = args[3]
	this.Receiver.Account = args[4]
	this.Receiver.Division = args[5]
	this.Security = args[6]
	this.Quantity = args[7]
	this.Reference = args[8]
	this.InstructionDate = args[9]
	this.TradeDate = args[10]
	//this.Reason					= args[11]

	return nil
}

func (this *Instruction) toLedgerValue() ([]byte, error) {
	return json.Marshal([]string{
		this.DeponentFrom, this.DeponentTo, this.Status, this.Initiator,
		this.AlamedaFrom, this.AlamedaTo, this.AlamedaSignatureFrom, this.AlamedaSignatureTo,
	})
}

func (this *Instruction) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Instruction) fillFromLedgerValue(bytes []byte) error {
	var str []string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	this.DeponentFrom = str[0]
	this.DeponentTo = str[1]
	this.Status = str[2]
	this.Initiator = str[3]
	this.AlamedaFrom = str[4]
	this.AlamedaTo = str[5]
	this.AlamedaSignatureFrom = str[6]
	this.AlamedaSignatureTo = str[7]

	return nil
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
		Initiator:      this.DeponentFrom,
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
	instructionWrapper.Initiator = this.DeponentTo

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
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	if authenticateCaller(stub, instruction.Receiver) == false {
		return shim.Error("TODO")
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
		//TODO change from 500 to bad request 400
		return shim.Error(err.Error())
	}

	if authenticateCaller(stub, instruction.Transferer) == false {
		return shim.Error("TODO")
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
	logger.Info("########### InstructionChaincode status ###########")
	logger.Info(args)

	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
	}

	status := args[len(args)-1]

	callerIsTransferer := authenticateCaller(stub, instruction.Transferer)
	callerIsReceiver := authenticateCaller(stub, instruction.Receiver)
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

	// TODO, REFACTORING: the following code is error prone as it tries to implements a state machine with a bunch of ifs
	if !(callerIsTransferer || callerIsReceiver || callerIsNSD) {
		return pb.Response{Status: 403}
	}
	if (callerIsTransferer || callerIsReceiver) && status != InstructionCanceled {
		return pb.Response{Status: 403}
	}
	if callerIsNSD && (status != InstructionDeclined || status != InstructionExecuted) {
		return pb.Response{Status: 403}
	}

	if instruction.existsIn(stub) {
		if err := instruction.loadFrom(stub); err != nil {
			return shim.Error(err.Error())
		}

		if instruction.Status == InstructionInitiated {
			if callerIsTransferer && instruction.Initiator != InitiatorIsTransferer {
				return pb.Response{Status: 403}
			}
			if callerIsReceiver && instruction.Initiator != InitiatorIsReceiver {
				return pb.Response{Status: 403}
			}
		}

		instruction.Status = status

		err = instruction.upsertIn(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		return pb.Response{Status: 404, Message: "instruction not found"}
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

		callerIsTransferer := authenticateCaller(stub, instruction.Transferer)
		callerIsReceiver := authenticateCaller(stub, instruction.Receiver)
		if !(callerIsTransferer || callerIsReceiver) {
			continue
		}

		if (callerIsTransferer && instruction.Initiator == InitiatorIsTransferer) ||
			(callerIsReceiver && instruction.Initiator == InitiatorIsReceiver) ||
			(instruction.Status == InstructionMatched) ||
			(instruction.Status == InstructionSigned) ||
			(instruction.Status == InstructionDeclined) {
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

func (t *InstructionChaincode) sign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting deponentFrom, accountFrom, divisionFrom, deponentTo, accountTo, divisionTo, " +
			"security, quantity, reference, instructionDate, tradeDate, signature")
	}

	signature := args[len(args)-1]

	instruction := Instruction{}
	err := instruction.fillFromArgs(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = instruction.loadFrom(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	callerIsTransferer := authenticateCaller(stub, instruction.Transferer)
	callerIsReceiver := authenticateCaller(stub, instruction.Receiver)

	if !(callerIsTransferer || callerIsReceiver) {
		return pb.Response{Status: 403}
	}

	if callerIsTransferer {
		instruction.AlamedaSignatureFrom = signature
	}

	if callerIsReceiver {
		instruction.AlamedaSignatureTo = signature
	}

	if instruction.AlamedaSignatureFrom != "" && instruction.AlamedaSignatureTo != "" {
		instruction.Status = InstructionSigned
	}

	if err := instruction.upsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	instruction.setEvent(stub)

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
		if bytes, err := stub.GetState(key); err == nil && getCreatorOrganization(stub) == string(bytes) {
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
