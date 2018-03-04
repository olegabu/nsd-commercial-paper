package nsd


import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)


// instruction statuses
type InstructionStatus string
const (
	InitiatorIsTransferer = "transferer"
	InitiatorIsReceiver   = "receiver"

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
const PositionIndex = `Position`






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

}

type InstructionValue struct {
	Status                  InstructionStatus `json:"status"`
	Initiator               string `json:"initiator"`

	DeponentFrom            string `json:"deponentFrom"`
	DeponentTo              string `json:"deponentTo"`
	MemberInstructionIdFrom string `json:"memberInstructionIdFrom"`
	MemberInstructionIdTo   string `json:"memberInstructionIdTo"`
	ReasonFrom              Reason `json:"reasonFrom"`
	ReasonTo                Reason `json:"reasonTo"`
	AlamedaFrom             string `json:"alamedaFrom"`
	AlamedaTo               string `json:"alamedaTo"`
	AlamedaSignatureFrom    string `json:"alamedaSignatureFrom"`
	AlamedaSignatureTo      string `json:"alamedaSignatureTo"`
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


// fill instruction key info
// return index of the first data field ( == key length)
func (this *Instruction) FillFromCompositeKeyParts(compositeKeyParts []string) (int, error) {
	keyLengthFop := 10;
	keyLengthDvp := keyLengthFop + 6;

	if len(compositeKeyParts) < keyLengthFop {
		return -1, errors.New("Composite key parts array length must be at least " + string(keyLengthFop))
	}

	if _, err := strconv.Atoi(compositeKeyParts[5]); err != nil {
		return -1, errors.New("Quantity must be int.")
	}

	this.Key.Transferer.Account  = compositeKeyParts[0]
	this.Key.Transferer.Division = compositeKeyParts[1]
	this.Key.Receiver.Account  = compositeKeyParts[2]
	this.Key.Receiver.Division = compositeKeyParts[3]
	this.Key.Security  = compositeKeyParts[4]
	this.Key.Quantity  = compositeKeyParts[5]
	this.Key.Reference = compositeKeyParts[6]
	this.Key.InstructionDate = compositeKeyParts[7]
	this.Key.TradeDate = compositeKeyParts[8]

	// TODO it's better to make Type the first parameter
	this.Key.Type = InstructionType(compositeKeyParts[9])
	if this.Key.Type != InstructionTypeFOP && this.Key.Type != InstructionTypeDVP {

		// support FOP by default
		this.Key.Type = InstructionTypeFOP
		return keyLengthFop - 1, nil
		//return -1, errors.New("Unknown instruction type " + string(this.Key.Type))
	}

	if this.Key.Type == InstructionTypeDVP {
		if len(compositeKeyParts) < keyLengthDvp {
			return keyLengthDvp, errors.New("Composite key parts array length must be at least " + string(keyLengthDvp))
		}

		// TODO: add DVP fields
		return keyLengthDvp, nil
	}

	return keyLengthFop, nil
}





func (this *Instruction) FillFromArgs(args []string) error {
	keyLength, err := this.FillFromCompositeKeyParts(args)
	if err != nil {
		return err
	}
	this.Key.Reference = strings.ToUpper(this.Key.Reference)

	//
	this.Value.DeponentFrom = args[ keyLength + 0 ]
	this.Value.DeponentTo   = args[ keyLength + 1 ]
	this.Value.MemberInstructionIdFrom = args[ keyLength + 2 ]
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