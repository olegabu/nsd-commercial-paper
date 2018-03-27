package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"
	"time"

	"bytes"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/olegabu/nsd-commercial-paper-common"
)

var logger = shim.NewLogger("InstructionChaincode")

const authenticationIndex = `Authentication`

// TODO: think about making these constants public in nsd.go
// args base lengths
const (
	fopArgsLength = 10
	dvpArgsLength = 16
)

type InstructionChaincode struct {
}

// **** Instruction Methods **** //

func matchIf(this *nsd.Instruction, stub shim.ChaincodeStubInterface,
			 desiredInitiator, desiredDeponentFrom, desiredDeponentTo string) pb.Response {
	if this.Value.Initiator != desiredInitiator {
		return pb.Response{Status: 400, Message: "Instruction is already created by " + this.Value.Initiator}
	}

	if this.Value.DeponentFrom != desiredDeponentFrom || this.Value.DeponentTo != desiredDeponentTo {
		return pb.Response{Status: 400, Message: "Deponents differ from entered by another party."}
	}

	if this.Value.Status != nsd.InstructionInitiated {
		return pb.Response{Status: 400, Message: "Instruction status is not " + nsd.InstructionInitiated}
	}

	this.Value.Status = nsd.InstructionMatched

	if this.Key.Type == nsd.InstructionTypeFOP {
		this.Value.AlamedaFrom, this.Value.AlamedaTo = createAlamedaFopXMLs(this)
	} else { // nsd.InstructionTypeDvp
		this.Value.AlamedaFrom, this.Value.AlamedaTo = createAlamedaDvpXMLs(this)
	}

	if err := this.UpsertIn(stub); err != nil {
		return pb.Response{Status: 500, Message: "Persistence failure."}
	}

	if err := this.EmitState(stub); err != nil {
		return pb.Response{Status: 500, Message: "Event emission failure."}
	}

	return shim.Success(nil)
}

func createAlamedaFopXMLs(this *nsd.Instruction) (string, string) {
	const xmlTemplate = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>{{.Depositary}}</deposit_c>
<contrag_c>{{.Initiator}}</contrag_c>
<contr_d_id>{{.InstructionID}}</contr_d_id>
<createdate>{{.Instruction.Key.InstructionDate}}</createdate>
<order_t_id>{{.OperationCode}}</order_t_id>
<execute_dt>{{.InstructionDate}}</execute_dt>
<expirat_dt>{{.ExpirationDate}}</expirat_dt>
</ORDER_HEADER>
<MF010>
<dep_acc_c>{{.Instruction.Key.Transferer.Account}}</dep_acc_c>
<sec_c>{{.Instruction.Key.Transferer.Division}}</sec_c>
<deponent_c>{{.Instruction.Value.DeponentFrom}}</deponent_c>
<corr_acc_c>{{.Instruction.Key.Receiver.Account}}</corr_acc_c>
<corr_sec_c>{{.Instruction.Key.Receiver.Division}}</corr_sec_c>
<corr_code>{{.Instruction.Value.DeponentTo}}</corr_code>{{if .ReasonExists}}{{with .Reason.Description -}}
<based_on>{{.}}</based_on>{{end}}
{{with .Reason.Document -}}<based_numb>{{.}}</based_numb>{{end}}
{{with .Reason.DocumentDate -}}<based_date>{{.}}</based_date>{{end}}{{end}}
<securities>
<security>
<security_c>{{.Instruction.Key.Security}}</security_c>
<security_q>{{.Instruction.Key.Quantity}}</security_q>
</security>
</securities>
<deal_reference>{{.Reference}}</deal_reference>
<date_deal>{{.Instruction.Key.TradeDate}}</date_deal>
</MF010>
</Document>
</Batch>
`

	type InstructionWrapper struct {
		Instruction     nsd.Instruction
		Depositary      string
		Initiator       string
		InstructionID   string
		OperationCode   string
		InstructionDate string
		ExpirationDate  string
		ReasonExists    bool
		Reason          nsd.Reason
		Reference       string
	}

	dateLayout := "2006-01-02"
	instructionDate, _ := time.Parse(dateLayout, this.Key.InstructionDate)
	expirationDate := instructionDate
	expirationDate = expirationDate.Truncate(time.Hour * 24).Add(time.Hour*(29*24+23) + time.Minute*59 + time.Second*59)

	instructionWrapper := InstructionWrapper{
		Instruction:     *this,
		Depositary:      "NDC000000000",
		Initiator:       this.Value.DeponentFrom,
		InstructionID:   this.Value.MemberInstructionIdFrom,
		OperationCode:   "16",
		InstructionDate: instructionDate.Format("2006-01-02 15:04:05"),
		ExpirationDate:  expirationDate.Format("2006-01-02 15:04:05"),
		Reason:          this.Value.ReasonFrom,
		Reference:       strings.ToUpper(this.Key.Reference),
	}
	instructionWrapper.ReasonExists = (instructionWrapper.Reason.Document != "") && (instructionWrapper.Reason.Description != "") && (instructionWrapper.Reason.DocumentDate != "")

	t := template.Must(template.New("xmlTemplate").Parse(xmlTemplate))

	buf := new(bytes.Buffer)
	t.Execute(buf, instructionWrapper)
	alamedaFrom := buf.String()

	buf.Reset()
	instructionWrapper.OperationCode = "16/1"
	instructionWrapper.Initiator = this.Value.DeponentTo
	instructionWrapper.InstructionID = this.Value.MemberInstructionIdTo
	instructionWrapper.Reason = this.Value.ReasonTo
	instructionWrapper.ReasonExists = (instructionWrapper.Reason.Document != "") && (instructionWrapper.Reason.Description != "") && (instructionWrapper.Reason.DocumentDate != "")

	t.Execute(buf, instructionWrapper)
	alamedaTo := buf.String()

	return alamedaFrom, alamedaTo
}

// TODO: get rid of test wrapper
func CreateAlamedaDvpXMLsTestWrapper(this *nsd.Instruction) (string, string) {
	return createAlamedaDvpXMLs(this)
}

func createAlamedaDvpXMLs(this *nsd.Instruction) (string, string) {
	const xmlTemplate = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>{{.Depositary}}</deposit_c>
<contrag_c>{{.Initiator}}</contrag_c>
<contr_d_id>{{.InstructionID}}</contr_d_id>
<createdate>{{.Instruction.Key.InstructionDate}}</createdate>
<order_t_id>{{.OperationCode}}</order_t_id>
<execute_dt>{{.InstructionDate}}</execute_dt>
<expirat_dt>{{.ExpirationDate}}</expirat_dt>
</ORDER_HEADER>
<MF170>
<dep_acc_c>{{.Instruction.Key.Transferer.Account}}</dep_acc_c>
<sec_c>{{.Instruction.Key.Transferer.Division}}</sec_c>
<corr_acc_c>{{.Instruction.Key.Receiver.Account}}</corr_acc_c>
<corr_sec_c>{{.Instruction.Key.Receiver.Division}}</corr_sec_c>
<deal_num>{{.Reference}}</deal_num>
<deal_date>{{.Instruction.Key.TradeDate}}</deal_date>
<con_code>{{.Contragent}}</con_code>
<sen_acc>{{.Instruction.Key.ReceiverRequisites.Account}}</sen_acc>
<sen_bic>{{.Instruction.Key.ReceiverRequisites.Bic}}</sen_bic>
<rec_acc>{{.Instruction.Key.TransfererRequisites.Account}}</rec_acc>
<rec_bic>{{.Instruction.Key.TransfererRequisites.Bic}}</rec_bic>
<pay_sum>{{.Instruction.Key.PaymentAmount}}</pay_sum>
<pay_curr>{{.Instruction.Key.PaymentCurrency}}</pay_curr>
{{if .ReasonExists}}{{with .Reason.Description -}}<based_on>{{.}}</based_on>{{end}}{{end}}
<block_securities>{{.BlockSecurities}}</block_securities>
<f_instruction>{{.FInstruction}}</f_instruction>
<auto_borr>{{.AutoBorr}}</auto_borr>{{if .AdditionalInfoExists}}
{{with .AdditionalInfo.Description -}}<add_info>{{.}}</add_info>{{end}}{{end}}
<securities>
<security>
<security_c>{{.Instruction.Key.Security}}</security_c>
<security_q>{{.Instruction.Key.Quantity}}</security_q>
</security>
</securities>
</MF170>
</Document>
</Batch>
`

	type InstructionWrapper struct {
		Instruction          nsd.Instruction
		Depositary           string
		Initiator      	     string
		InstructionID        string
		OperationCode        string
		InstructionDate      string
		ExpirationDate       string
		ReasonExists    	 bool
		Reason          	 nsd.Reason
		Reference       	 string
		Contragent      	 string
		BlockSecurities      string
		FInstruction         string
		AutoBorr             string
		AdditionalInfoExists bool
		AdditionalInfo       nsd.Reason
	}

	dateLayout := "2006-01-02"
	instructionDate, _ := time.Parse(dateLayout, this.Key.InstructionDate)
	expirationDate := instructionDate
	expirationDate = expirationDate.Truncate(time.Hour * 24).Add(time.Hour*23 + time.Minute*59 + time.Second*59)

	instructionWrapper := InstructionWrapper{
		Instruction:          *this,
		Depositary:           "NDC000000000",
		Initiator:            this.Value.DeponentFrom,
		InstructionID:        this.Value.MemberInstructionIdFrom,
		OperationCode:        "16/2",
		InstructionDate:      instructionDate.Format("2006-01-02 15:04:05"),
		ExpirationDate:       expirationDate.Format("2006-01-02 15:04:05"),
		Reason:               this.Value.ReasonFrom,
		Reference:            strings.ToUpper(this.Key.Reference),
		Contragent:           this.Value.DeponentTo,
		BlockSecurities:      "N",
		FInstruction:         "N",
		AutoBorr:             "N",
		AdditionalInfoExists: false,
		AdditionalInfo:       this.Value.AdditionalInformation,
	}
	instructionWrapper.ReasonExists = (instructionWrapper.Reason.Document != "") &&
		(instructionWrapper.Reason.Description != "") && (instructionWrapper.Reason.DocumentDate != "")

	t := template.Must(template.New("xmlTemplate").Parse(xmlTemplate))

	buf := new(bytes.Buffer)
	t.Execute(buf, instructionWrapper)
	alamedaFrom := buf.String()

	buf.Reset()
	instructionWrapper.OperationCode = "16/3"
	instructionWrapper.Initiator = this.Value.DeponentTo
	instructionWrapper.InstructionID = this.Value.MemberInstructionIdTo
	instructionWrapper.Reason = this.Value.ReasonTo
	instructionWrapper.ReasonExists = (instructionWrapper.Reason.Document != "") &&
		(instructionWrapper.Reason.Description != "") && (instructionWrapper.Reason.DocumentDate != "")
	instructionWrapper.Contragent = this.Value.DeponentFrom
	instructionWrapper.FInstruction = "Y"
	instructionWrapper.AdditionalInfoExists = true
	instructionWrapper.AdditionalInfo.Description = "/NZP " + instructionWrapper.AdditionalInfo.Description

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
		Name     string        `json:"organization"`
		Balances []nsd.Balance `json:"balances"`
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
	} else {
		return pb.Response{Status: 400, Message: "JSON unmarshalling error."}
	}

	return shim.Success(nil)
}

func (t *InstructionChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### InstructionChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "receive" {
		if len(args) < fopArgsLength + 4 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.receive(stub, args)
	}
	if function == "transfer" {
		if len(args) < fopArgsLength + 4 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.transfer(stub, args)
	}
	if function == "status" {
		if len(args) < fopArgsLength + 1 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.status(stub, args)
	}
	if function == "query" {
		if len(args) < 0 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.query(stub, args)
	}
	if function == "queryByType" {
		if len(args) < 1 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.queryByType(stub, args)
	}
	if function == "history" {
		if len(args) < fopArgsLength {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.history(stub, args)
	}
	if function == "sign" {
		if len(args) < fopArgsLength + 1 {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.sign(stub, args)
	}
	if function == "rollback" {
		if len(args) < fopArgsLength {
			return pb.Response{Status: 400, Message: "Incorrect number of arguments."}
		}
		return t.rollback(stub, args)
	}

	err := fmt.Sprintf("Unknown function, check the first argument, must be one of: "+
		"receive, transfer, query, history, status, sign, rollback. But got: %v", args[0])
	logger.Error(err)
	return shim.Error(err)
}

func isInstructionUnique(stub shim.ChaincodeStubInterface, instruction nsd.Instruction) (bool, error) {
	const instructionUnique = true

	it, err := stub.GetStateByPartialCompositeKey(nsd.InstructionIndex, []string{})
	if err != nil {
		return instructionUnique, err
	}
	defer it.Close()

	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return instructionUnique, err
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return instructionUnique, err
		}

		ledgerInstruction := nsd.Instruction{}

		if err := ledgerInstruction.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			return instructionUnique, err
		}

		if ledgerInstruction.Key.Reference == instruction.Key.Reference &&
			ledgerInstruction.Key.InstructionDate == instruction.Key.InstructionDate {
			return !instructionUnique, nil
		}
	}

	return instructionUnique, nil
}

func (t *InstructionChaincode) receive(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	if authenticateCaller(stub, instruction.Key.Receiver) == false {
		return pb.Response{Status: 403, Message: "Caller must be receiver."}
	}

	var argsOffset int
	if instruction.Key.Type == nsd.InstructionTypeFOP {
		argsOffset = len(args) - 4
	} else { // nsd.InstructionTypeDVP
		argsOffset = len(args) - 5
	}

	if instruction.ExistsIn(stub) {
		if err := instruction.LoadFrom(stub); err != nil {
			return pb.Response{Status: 404, Message: "Instruction not found."}
		}

		instruction.Value.MemberInstructionIdTo = args[argsOffset + 2]
		if err := json.Unmarshal([]byte(args[argsOffset + 3]), &instruction.Value.ReasonTo); err != nil {
			return pb.Response{Status: 400, Message: "Wrong arguments."}
		}

		if instruction.Key.Type == nsd.InstructionTypeDVP {
			// additional info argument passed
			if err := json.Unmarshal([]byte(args[argsOffset + 4]), &instruction.Value.AdditionalInformation);
			   err != nil {
				return pb.Response{Status: 400, Message: "Wrong arguments."}
			}
		}

		if instruction.UpsertIn(stub) != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}

		return matchIf(&instruction, stub, nsd.InitiatorIsTransferer, args[argsOffset], args[argsOffset + 1])
	} else {
		instruction.Value.DeponentFrom = args[argsOffset]
		instruction.Value.DeponentTo = args[argsOffset + 1]
		instruction.Value.MemberInstructionIdTo = args[argsOffset + 2]
		instruction.Value.Initiator = nsd.InitiatorIsReceiver
		instruction.Value.Status = nsd.InstructionInitiated
		if err := json.Unmarshal([]byte(args[argsOffset + 3]), &instruction.Value.ReasonTo); err != nil {
			return pb.Response{Status: 400, Message: "Wrong arguments."}
		}
		if instruction.Key.Type == nsd.InstructionTypeDVP {
			// additional info argument passed
			if err := json.Unmarshal([]byte(args[argsOffset + 4]), &instruction.Value.AdditionalInformation);
				err != nil {
				return pb.Response{Status: 400, Message: "Wrong arguments."}
			}
		}

		isUnique, err := isInstructionUnique(stub, instruction)
		if err != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}
		if !isUnique {
			return pb.Response{Status: 400, Message: "Instruction is not unique."}
		}

		if instruction.UpsertIn(stub) != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}

		}
		return shim.Success(nil)
	}
}

func (t *InstructionChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	if authenticateCaller(stub, instruction.Key.Transferer) == false {
		return pb.Response{Status: 403, Message: "Caller must be transferer."}
	}

	argsOffset := len(args) - 4

	if instruction.ExistsIn(stub) {
		if err := instruction.LoadFrom(stub); err != nil {
			return pb.Response{Status: 404, Message: "Instruction not found."}
		}

		instruction.Value.MemberInstructionIdFrom = args[argsOffset + 2]
		if err := json.Unmarshal([]byte(args[argsOffset + 3]), &instruction.Value.ReasonFrom); err != nil {
			return pb.Response{Status: 400, Message: "Wrong arguments."}
		}

		if instruction.UpsertIn(stub) != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}

		}
		return matchIf(&instruction, stub, nsd.InitiatorIsReceiver, args[argsOffset], args[argsOffset + 1])
	} else {
		instruction.Value.DeponentFrom = args[argsOffset]
		instruction.Value.DeponentTo = args[argsOffset + 1]
		instruction.Value.MemberInstructionIdFrom = args[argsOffset + 2]
		instruction.Value.Initiator = nsd.InitiatorIsTransferer
		instruction.Value.Status = nsd.InstructionInitiated
		if err := json.Unmarshal([]byte(args[argsOffset + 3]), &instruction.Value.ReasonFrom); err != nil {
			return pb.Response{Status: 400, Message: "Wrong arguments."}
		}

		isUnique, err := isInstructionUnique(stub, instruction)
		if err != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}
		if !isUnique {
			return pb.Response{Status: 400, Message: "Instruction is not unique."}
		}

		if instruction.UpsertIn(stub) != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}

		}
		return shim.Success(nil)
	}
}

func (t *InstructionChaincode) status(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("########### InstructionChaincode status ###########")
	logger.Info(args)

	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
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

	if err := instruction.LoadFrom(stub); err != nil {
		return pb.Response{Status: 404, Message: "Instruction not found."}
	}

	var expectedArgsLength int
	if instruction.Key.Type == nsd.InstructionTypeFOP {
		expectedArgsLength = fopArgsLength + 1
	} else { // nsd.InstructionTypeDVP
		expectedArgsLength = dvpArgsLength + 1
	}

	if len(args) > expectedArgsLength {
		instruction.Value.StatusInfo = args[len(args) - 2]
	}

	switch {
	case callerIsNSD && status == nsd.InstructionDeclined,
		 callerIsNSD && status == nsd.InstructionExecuted,
		 callerIsNSD && status == nsd.InstructionDownloaded,
		 callerIsNSD && status == nsd.InstructionRollbackInitiated,
		 callerIsNSD && status == nsd.InstructionRollbackDone,
		 callerIsNSD && status == nsd.InstructionRollbackDeclined:
		instruction.Value.Status = status
		if err := instruction.UpsertIn(stub); err != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}
	case (callerIsTransferer || callerIsReceiver) && instruction.Value.Status == nsd.InstructionInitiated &&
		 status == nsd.InstructionCanceled:
		if (callerIsTransferer && instruction.Value.Initiator == nsd.InitiatorIsTransferer) ||
			(callerIsReceiver && instruction.Value.Initiator == nsd.InitiatorIsReceiver) {
			instruction.Value.Status = status
			if err := instruction.UpsertIn(stub); err != nil {
				return pb.Response{Status: 500, Message: "Persistence failure."}
			}
		}
	default:
		return pb.Response{Status: 406, Message: "Instruction status or caller identity is wrong."}
	}

	if err := instruction.EmitState(stub); err != nil {
		return pb.Response{Status: 500, Message: "Event emission failure."}
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

//TODO: move this code to common package
func (t *InstructionChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(nsd.InstructionIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	instructions := []nsd.Instruction{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction := nsd.Instruction{}

		if err := instruction.FillFromLedgerValue(response.Value); err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		if err := instruction.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			return shim.Error(err.Error())
		}

		callerIsTransferer := authenticateCaller(stub, instruction.Key.Transferer)
		callerIsReceiver := authenticateCaller(stub, instruction.Key.Receiver)
		callerIsNSD := getCreatorOrganization(stub) == "nsd.nsd.ru"

		logger.Debug(callerIsTransferer, callerIsReceiver, callerIsNSD)

		if !(callerIsTransferer || callerIsReceiver || callerIsNSD) {
			continue
		}

		if (callerIsTransferer && instruction.Value.Initiator == nsd.InitiatorIsTransferer) ||
			(callerIsReceiver && instruction.Value.Initiator == nsd.InitiatorIsReceiver) ||
			(instruction.Value.Status == nsd.InstructionMatched) ||
			(instruction.Value.Status == nsd.InstructionSigned) ||
			(instruction.Value.Status == nsd.InstructionExecuted) ||
			(instruction.Value.Status == nsd.InstructionDownloaded) ||
			(instruction.Value.Status == nsd.InstructionDeclined) ||
			(instruction.Value.Status == nsd.InstructionRollbackInitiated) ||
			(instruction.Value.Status == nsd.InstructionRollbackDone) ||
			(instruction.Value.Status == nsd.InstructionRollbackDeclined) ||
			callerIsNSD {
			instructions = append(instructions, instruction)
		}
	}

	result, err := json.Marshal(instructions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *InstructionChaincode) queryByType(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// status
	if len(args) != 1 {
		return pb.Response{Status: 400, Message: fmt.Sprintf("Incorrect number of arguments. "+
			"Expecting 'status'. "+
			"But got %d args: %s", len(args), args)}
	}

	expectedStatus := args[0]

	it, err := stub.GetStateByPartialCompositeKey(nsd.InstructionIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	instructions := []nsd.Instruction{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		instruction := nsd.Instruction{}

		if err := instruction.FillFromLedgerValue(response.Value); err != nil {
			return shim.Error(err.Error())
		}

		if instruction.Value.Status == expectedStatus {
			instructions = append(instructions, instruction)
		}
	}

	result, err := json.Marshal(instructions)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

//TODO: move this code to common package
func (t *InstructionChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
	}

	compositeKey, err := instruction.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	it, err := stub.GetHistoryForKey(compositeKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	modifications := []nsd.InstructionHistoryValue{}

	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var entry nsd.InstructionHistoryValue

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
	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "cannot initialize instruction from args"}
	}

	signature := args[len(args)-1]

	if err := instruction.LoadFrom(stub); err != nil {
		return pb.Response{Status: 404, Message: "Instruction not found."}
	}

	callerIsTransferer := authenticateCaller(stub, instruction.Key.Transferer)
	callerIsReceiver := authenticateCaller(stub, instruction.Key.Receiver)

	if !(callerIsTransferer || callerIsReceiver) {
		return pb.Response{Status: 403, Message: "Caller must be either transferer or receiver."}
	}

	if callerIsTransferer {
		instruction.Value.AlamedaSignatureFrom = signature
	}

	if callerIsReceiver {
		instruction.Value.AlamedaSignatureTo = signature
	}

	if instruction.Value.AlamedaSignatureFrom != "" && instruction.Value.AlamedaSignatureTo != "" {
		instruction.Value.Status = nsd.InstructionSigned
		instruction.EmitState(stub)
	}

	if err := instruction.UpsertIn(stub); err != nil {
		return pb.Response{Status: 500, Message: "Persistence failure."}
	}

	return shim.Success(nil)
}

func (t *InstructionChaincode) rollback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	callerIsNSD := getCreatorOrganization(stub) == "nsd.nsd.ru"

	if !callerIsNSD {
		return pb.Response{Status: 403, Message: "Instruction can be rolled back only by NSD."}
	}

	instruction := nsd.Instruction{}
	if err := instruction.FillFromArgs(args); err != nil {
		return pb.Response{Status: 400, Message: "Wrong arguments."}
	}

	if instruction.ExistsIn(stub) {
		if err := instruction.LoadFrom(stub); err != nil {
			return pb.Response{Status: 404, Message: "Instruction not found."}
		}

		instruction.Value.Status = nsd.InstructionRollbackInitiated

		if instruction.UpsertIn(stub) != nil {
			return pb.Response{Status: 500, Message: "Persistence failure."}
		}
		if err := instruction.EmitState(stub); err != nil {
			return pb.Response{Status: 500, Message: "Event emission failure."}
		}

		return shim.Success(nil)
	} else {
		return pb.Response{Status: 404, Message: "Instruction does not exist in ledger."}
	}
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

func authenticateCaller(stub shim.ChaincodeStubInterface, callerBalance nsd.Balance) bool {
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
