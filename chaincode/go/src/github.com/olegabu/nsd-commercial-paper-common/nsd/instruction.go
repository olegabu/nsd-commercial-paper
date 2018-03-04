package nsd

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/olegabu/nsd-commercial-paper-common/assert"
	"strconv"
)

type InstructionInitiator string
const (
	InitiatorIsTransferer InstructionInitiator = "transferer"
	InitiatorIsReceiver   InstructionInitiator = "receiver"
)

// instruction statuses
type InstructionStatus string
const (
	InstructionInitiated  InstructionStatus = "initiated"
	InstructionMatched    InstructionStatus = "matched"
	InstructionSigned     InstructionStatus = "signed"
	InstructionExecuted   InstructionStatus = "executed"
	InstructionDownloaded InstructionStatus = "downloaded"
	InstructionDeclined   InstructionStatus = "declined"
	InstructionCanceled   InstructionStatus = "canceled"
)

// instruction types
type InstructionType string
const (
	InstructionTypeFOP InstructionType = "fop"
	InstructionTypeDVP InstructionType = "dvp"
)

// TODO: make this private
const InstructionIndex = `Instruction`

// Instruction is the main data type stored in ledger
type Instruction struct {
	Key   InstructionKey   `json:"key"`
	Value InstructionValue `json:"value"`
}

type InstructionKey struct {
	Transferer Balance `json:"transferer"`
	Receiver   Balance `json:"receiver"`
	Security   string  `json:"security"`
	//TODO Quantity should be int like everywhere
	Quantity        string          `json:"quantity"`
	Reference       string          `json:"reference"`
	InstructionDate string          `json:"instructionDate"`
	TradeDate       string          `json:"tradeDate"`
	Type            InstructionType `json:"type"`

	// Payment* fields is filled only for Type==DVP
	// if Type==FOP these fields is an empty string

	// PaymentFrom is logically 'Receiver'
	PaymentFrom InstructionPaymentCounteragent `json:"paymentFrom"`
	// PaymentTo is logically 'Transferer'
	PaymentTo InstructionPaymentCounteragent `json:"paymentTo"`
	//TODO PaymentAmount should be int like everywhere
	PaymentAmount   string `json:"paymentAmount"`
	PaymentCurrency string `json:"paymentCurrency"`
}

type InstructionValue struct {
	//CounteragentFrom InstructionContragent `json:"from"`
	//CounteragentTo   InstructionContragent `json:"to"`

	Status    InstructionStatus    `json:"status"`
	Initiator InstructionInitiator `json:"initiator"`

	DeponentFrom            string `json:"deponentFrom"`
	DeponentTo              string `json:"deponentTo"`

	// TODO: Use CounteragentFrom/CounteragentTo
	MemberInstructionIdFrom string `json:"memberInstructionIdFrom"`
	MemberInstructionIdTo   string `json:"memberInstructionIdTo"`

	ReasonFrom              Reason `json:"reasonFrom"`
	ReasonTo                Reason `json:"reasonTo"`
	AlamedaFrom             string `json:"alamedaFrom"`
	AlamedaTo               string `json:"alamedaTo"`
	AlamedaSignatureFrom    string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo      string `json:"alamedaSignatureTo"`
}

type InstructionContragent struct {
	Deponent            string `json:"deponent"`
	MemberInstructionId string `json:"memberInstructionId"`
	Reason              Reason `json:"reason"`
	Alameda             string `json:"alameda"`
	AlamedaSignature    string `json:"alamedaSignature"`
}

type InstructionPaymentCounteragent struct {
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

// fill instruction key info by compositeKey
func (this *Instruction) FromCompositeKey(stub shim.ChaincodeStubInterface, compositeKey string) (error) {
	_, compositeKeyParts, err := stub.SplitCompositeKey(compositeKey)
	if err != nil {
		return err
	}

	if err := this.FillKeyFromArgs(compositeKeyParts); err != nil {
		return err
	}
	return nil
}

// get compositeKey for the instruction
func (this *Instruction) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	keyParts := []string{
		string(this.Key.Type),

		this.Key.Transferer.Account,
		this.Key.Transferer.Division,
		this.Key.Receiver.Account,
		this.Key.Receiver.Division,
		this.Key.Security,
		this.Key.Quantity,
		this.Key.Reference,
		this.Key.InstructionDate,
		this.Key.TradeDate,

		this.Key.PaymentFrom.Account,
		this.Key.PaymentFrom.Bic,

		this.Key.PaymentTo.Account,
		this.Key.PaymentTo.Bic,

		this.Key.PaymentAmount,
		this.Key.PaymentCurrency,
	}
	return stub.CreateCompositeKey(InstructionIndex, keyParts)
}



// FillFromCompositeKeyParts() is the opposite to ToCompositeKey()
func (this *Instruction) FillKeyFromArgs(compositeKeyParts []string) (error) {
	keyLengthFop := 10
	keyLengthDvp := keyLengthFop + 6

	if len(compositeKeyParts) < keyLengthFop {
		return errors.New("Composite key parts array length must be at least " + strconv.Itoa(keyLengthFop))
	}

	fieldOffset := 0

	// get and check instruction type
	this.Key.Type = InstructionType(compositeKeyParts[fieldOffset+0])

	if this.Key.Type != InstructionTypeFOP && this.Key.Type != InstructionTypeDVP {
		// FOP by default
		this.Key.Type = InstructionTypeFOP
		fieldOffset = -1
		//return -1, errors.New("Unknown instruction type " + string(this.Key.Type))
	}

	// check arguments length
	if this.Key.Type == InstructionTypeDVP {
		if len(compositeKeyParts) < keyLengthDvp {
			return errors.New("Composite key parts array length must be at least " + strconv.Itoa(keyLengthDvp))
		}
	} else if this.Key.Type == InstructionTypeFOP {
		if len(compositeKeyParts) < keyLengthFop {
			return errors.New("Composite key parts array length must be at least " + strconv.Itoa(keyLengthFop))
		}
	} else {
		return errors.New("Invalid instruction type: " + string(this.Key.Type))
	}

	// this.Key.Quantity
	if !assert.IsNumber(compositeKeyParts[fieldOffset+6]) {
		return errors.New("Quantity must be int.")
	}

	this.Key.Transferer.Account 	= compositeKeyParts[fieldOffset+1]
	this.Key.Transferer.Division 	= compositeKeyParts[fieldOffset+2]
	this.Key.Receiver.Account 		= compositeKeyParts[fieldOffset+3]
	this.Key.Receiver.Division 		= compositeKeyParts[fieldOffset+4]
	this.Key.Security 				= compositeKeyParts[fieldOffset+5]
	this.Key.Quantity 				= compositeKeyParts[fieldOffset+6]
	this.Key.Reference 				= compositeKeyParts[fieldOffset+7]
	this.Key.InstructionDate 		= compositeKeyParts[fieldOffset+8]
	this.Key.TradeDate 				= compositeKeyParts[fieldOffset+9]

	if this.Key.Type == InstructionTypeDVP {
		this.Key.PaymentFrom.Account 	= compositeKeyParts[fieldOffset+10]
		this.Key.PaymentFrom.Bic 		= compositeKeyParts[fieldOffset+11]
		this.Key.PaymentTo.Account 		= compositeKeyParts[fieldOffset+12]
		this.Key.PaymentTo.Bic 			= compositeKeyParts[fieldOffset+13]
		this.Key.PaymentAmount 			= compositeKeyParts[fieldOffset+14]
		this.Key.PaymentCurrency		= compositeKeyParts[fieldOffset+15]
	}

	this.Key.Reference = strings.ToUpper(this.Key.Reference)
	return nil
}

//
//
func (this *Instruction) FillValueFromArgs(args []string, initiator InstructionInitiator) error {
	fieldOffset := len(args) - 4 // get last 4 arguments

	this.Value.Initiator = initiator

	//
	this.Value.DeponentFrom = args[fieldOffset+0]
	this.Value.DeponentTo = args[fieldOffset+1]

	// parse reason
	reason := Reason{};
	if err := json.Unmarshal([]byte(args[fieldOffset+3]), &reason); err != nil {
		return err
	}

	if initiator == InitiatorIsTransferer {
		this.Value.MemberInstructionIdFrom = args[fieldOffset+2]
		this.Value.ReasonFrom = reason;
	} else if initiator == InitiatorIsReceiver {
		this.Value.MemberInstructionIdTo = args[fieldOffset+2]
		this.Value.ReasonTo = reason;
	} else {
		return errors.New("Invalid initiator: " + string(initiator))
	}

	return nil
}


func (this *Instruction) FillFromArgs(args []string, initiator InstructionInitiator) error {
	if err := this.FillKeyFromArgs(args); err != nil {
		return err
	}
	if err := this.FillValueFromArgs(args, initiator); err != nil {
		return err
	}
	return nil
}

//
//
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

//
//
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

//
//
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

	if err = stub.SetEvent(InstructionIndex+"."+string(this.Value.Status), data); err != nil {
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
