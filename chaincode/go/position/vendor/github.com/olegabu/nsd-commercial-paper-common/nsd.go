package nsd

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"

	InstructionInitiated  = "initiated"
	InstructionMatched    = "matched"
	InstructionSigned     = "signed"
	InstructionExecuted   = "executed"
	InstructionDownloaded = "downloaded"
	InstructionDeclined   = "declined"
	InstructionCanceled   = "canceled"
)

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
