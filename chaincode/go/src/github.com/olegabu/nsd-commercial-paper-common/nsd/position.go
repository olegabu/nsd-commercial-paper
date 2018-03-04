package nsd

import (
	"errors"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
)

// Position is the main data type stored in ledger
type Position struct {
	Balance  Balance `json:"balance"`
	Security string  `json:"security"`
	Quantity int     `json:"quantity"`
}

const PositionIndex = `Position`


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

