package nsd

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// Instruction initiators
const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"
)

// Instruction statuses
const (
	InstructionInitiated  = "initiated"
	InstructionMatched    = "matched"
	InstructionSigned     = "signed"
	InstructionExecuted   = "executed"
	InstructionDownloaded = "downloaded"
	InstructionDeclined   = "declined"
	InstructionCanceled   = "canceled"

	InstructionRollbackInitiated = "rollbackInitiated"
	InstructionRollbackDone      = "rollbackDone"
	InstructionRollbackDeclined  = "rollbackDeclined"
)

// Instruction types
const (
	InstructionTypeFOP = "fop"
	InstructionTypeDVP = "dvp"
)

// Args lengths
const fopArgsLength = 10
const dvpArgsLength = 16

// TODO: make this private
const InstructionIndex = `Instruction`
const PositionIndex = `Position`

// Instruction is the main data type stored in ledger
type Instruction struct {
	Key   InstructionKey   `json:"key"`
	Value InstructionValue `json:"value"`
}

// Position is the main data type stored in ledger
type Position struct {
	Balance  Balance `json:"balance"`
	Security string  `json:"security"`
	Quantity int     `json:"quantity"`
}

type InstructionKey struct {
	Transferer Balance `json:"transferer"`
	Receiver   Balance `json:"receiver"`

	Security        string `json:"security"`
	// TODO: quantity should be int like everywhere
	Quantity        string `json:"quantity"`
	Reference       string `json:"reference"`
	InstructionDate string `json:"instructionDate"`
	TradeDate       string `json:"tradeDate"`

	// valid types are described by "Instruction types" constants
	Type string `json:"type"`

	// block below is used only if Type == InstructionTypeDVP
	TransfererRequisites Requisites `json:"transfererRequisites"`
	ReceiverRequisites   Requisites `json:"receiverRequisites"`
	// TODO: amount should be float
	PaymentAmount        string     `json:"paymentAmount"`
	PaymentCurrency      string     `json:"paymentCurrency"`
}

type InstructionValue struct {
	DeponentFrom                  string `json:"deponentFrom"`
	DeponentTo                    string `json:"deponentTo"`
	Status                        string `json:"status"`
	StatusInfo                    string `json:"statusInfo"`
	Initiator                     string `json:"initiator"`
	MemberInstructionIdFrom       string `json:"memberInstructionIdFrom"`
	MemberInstructionIdTo         string `json:"memberInstructionIdTo"`
	ReasonFrom                    Reason `json:"reasonFrom"`
	ReasonTo                      Reason `json:"reasonTo"`
	AlamedaFrom                   string `json:"alamedaFrom"`
	AlamedaTo                     string `json:"alamedaTo"`
	AlamedaSignatureFrom          string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo            string `json:"alamedaSignatureTo"`
	ReceiverSignatureDownloaded   bool   `json:"receiverSignatureDownloaded"`
	TransfererSignatureDownloaded bool   `json:"transfererSignatureDownloaded"`
	AdditionalInformation         Reason `json:"additionalInformation"`
}

type Balance struct {
	Account  string `json:"account"`
	Division string `json:"division"`
}

type Requisites struct {
	Account string `json:"account"`
	Bic     string `json:"bic"`
}

type Reason struct {
	DocumentDate string `json:"created"`
	Document     string `json:"document"`
	Description  string `json:"description"`
}

// required for history
type InstructionHistoryValue struct {
	TxId      string           `json:"txId"`
	Value     InstructionValue `json:"value"`
	Timestamp string           `json:"timestamp"`
	IsDelete  bool             `json:"isDelete"`
}

func (this *Instruction) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
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
		this.Key.Type,
		this.Key.TransfererRequisites.Account,
		this.Key.TransfererRequisites.Bic,
		this.Key.ReceiverRequisites.Account,
		this.Key.ReceiverRequisites.Bic,
		this.Key.PaymentAmount,
		this.Key.PaymentCurrency,
	}
	return stub.CreateCompositeKey(InstructionIndex, keyParts)
}

func (this *Instruction) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < fopArgsLength {
		return errors.New("Composite key parts array length must be at least 9.")
	}

	if _, err := strconv.Atoi(compositeKeyParts[5]); err != nil {
		return errors.New("Quantity must be int.")
	}

	if compositeKeyParts[9] != InstructionTypeFOP && compositeKeyParts[9] != InstructionTypeDVP {
		return errors.New("Type of instruction must be either \"fop\" or \"dvp\".")
	}

	if compositeKeyParts[9] == InstructionTypeDVP {
		if len(compositeKeyParts) < dvpArgsLength {
			return errors.New("Composite key parts array length for \"dvp\" option must be at least 16.")
		}

		if _, err := strconv.ParseFloat(compositeKeyParts[14], 64); err != nil {
			return errors.New("Payment amount must be float (dvp).")
		}

		this.Key.TransfererRequisites.Account = compositeKeyParts[10]
		this.Key.TransfererRequisites.Bic = compositeKeyParts[11]
		this.Key.ReceiverRequisites.Account = compositeKeyParts[12]
		this.Key.ReceiverRequisites.Bic = compositeKeyParts[13]
		this.Key.PaymentAmount = compositeKeyParts[14]
		this.Key.PaymentCurrency = compositeKeyParts[15]
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
	this.Key.Type = compositeKeyParts[9]

	return nil
}

func (this *Instruction) FillFromArgs(args []string) error {
	if err := this.FillFromCompositeKeyParts(args); err != nil {
		return err
	}
	this.Key.Reference = strings.ToUpper(this.Key.Reference)
	return nil
}

func (this *Instruction) ExistsIn(stub shim.ChaincodeStubInterface) bool {
	compositeKey, err := this.ToCompositeKey(stub)
	if err != nil {
		return false
	}

	if data, err := stub.GetState(compositeKey); err != nil || data == nil {
		return false
	}

	return true
}

func (this *Instruction) LoadFrom(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	data, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return this.FillFromLedgerValue(data)
}

func (this *Instruction) UpsertIn(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.ToCompositeKey(stub)
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

func (this *Instruction) EmitState(stub shim.ChaincodeStubInterface) error {
	data, err := this.toJSON()
	if err != nil {
		return err
	}

	if err = stub.SetEvent(InstructionIndex+"."+this.Value.Status, data); err != nil {
		return err
	}

	return nil
}

func (this *Instruction) toLedgerValue() ([]byte, error) {
	return json.Marshal(this.Value)
}

func (this *Instruction) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Instruction) FillFromLedgerValue(bytes []byte) error {
	if err := json.Unmarshal(bytes, &this.Value); err != nil {
		return err
	} else {
		return nil
	}
}

// **** Position Methods **** //
func (this *Position) toStringArray() []string {
	return []string{
		this.Balance.Account,
		this.Balance.Division,
		this.Security,
	}
}

func (this *Position) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(PositionIndex, this.toStringArray())
}

func (this *Position) existsIn(stub shim.ChaincodeStubInterface) bool {
	compositeKey, err := this.ToCompositeKey(stub)
	if err != nil {
		return false
	}

	bytes, _ := stub.GetState(compositeKey)
	if bytes == nil {
		return false
	}

	return true
}

func (this *Position) loadFrom(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	bytes, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return this.FillFromLedgerValue(bytes)
}

func (this *Position) UpsertIn(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := this.ToCompositeKey(stub)
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

func (this *Position) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) != 3 {
		return errors.New("Composite key parts array length must be 3.")
	}

	this.Balance.Account = compositeKeyParts[0]
	this.Balance.Division = compositeKeyParts[1]
	this.Security = compositeKeyParts[2]

	return nil
}

func (this *Position) FillFromArgs(args []string) error {
	if len(args) < 3 {
		return errors.New("Incorrect number of arguments. Expecting >=3.")
	}

	this.Balance.Account = args[0]
	this.Balance.Division = args[1]
	this.Security = args[2]

	if len(args) > 3 {
		quantity, err := strconv.Atoi(args[3])
		if err != nil {
			return errors.New("cannot convert to quantity")
		}
		this.Quantity = quantity
	}

	return nil
}

func (this *Position) toLedgerValue() ([]byte, error) {
	return json.Marshal([]string{strconv.Itoa(this.Quantity)})
}

func (this *Position) toJSON() ([]byte, error) {
	return json.Marshal(this)
}

func (this *Position) FillFromLedgerValue(bytes []byte) error {
	var str []string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	quantity, err := strconv.Atoi(str[0])
	if err != nil {
		return err
	}
	this.Quantity = quantity

	return nil
}
