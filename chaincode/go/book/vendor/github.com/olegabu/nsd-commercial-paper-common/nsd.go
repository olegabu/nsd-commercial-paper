package nsd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ************************************** TODO: TO BE MOVED TO COMMON PACKAGE ************************************** //
const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"

	InstructionInitiated    = "initiated"
	InstructionMatched      = "matched"
	InstructionSigned       = "signed"
	InstructionExecuted     = "executed"
	InstructionDownloaded   = "downloaded"
	InstructionDeclined     = "declined"
	InstructionCanceled     = "canceled"
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
	//TODO should be int like everywhere
	Quantity        string `json:"quantity"`
	Reference       string `json:"reference"`
	InstructionDate string `json:"instructionDate"`
	TradeDate       string `json:"tradeDate"`
}

type InstructionValue struct {
	DeponentFrom            string `json:"deponentFrom"`
	DeponentTo              string `json:"deponentTo"`
	Status                  string `json:"status"`
	Initiator               string `json:"initiator"`
	MemberInstructionIdFrom string `json:"memberInstructionIdFrom"`
	MemberInstructionIdTo   string `json:"memberInstructionIdTo"`
	ReasonFrom              Reason `json:"reasonFrom"`
	ReasonTo                Reason `json:"reasonTo"`
	AlamedaFrom             string `json:"alamedaFrom"`
	AlamedaTo               string `json:"alamedaTo"`
	AlamedaSignatureFrom    string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo      string `json:"alamedaSignatureTo"`
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
	}
	return stub.CreateCompositeKey(InstructionIndex, keyParts)
}

func (this *Instruction) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < 9 {
		return errors.New("Composite key parts array length must be at least 9.")
	}

	if _, err := strconv.Atoi(compositeKeyParts[5]); err != nil {
		return errors.New("Quantity must be int.")
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

func (this *Instruction) FillFromArgs(args []string) error {
	return this.FillFromCompositeKeyParts(args)
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

// ***************************************************************************************************************** //
