package nsd

import (
	"encoding/json"
)


type Balance struct {
	Account  string `json:"account"`
	Division string `json:"division"`
}


type Organization struct {
	Name     string    `json:"organization"`
	Deponent string    `json:"deponent"`
	Balances []Balance `json:"balances"`
}


func (this *Organization) ToJSON() []byte {
	r, err := json.Marshal(this)
	if err != nil {
		return nil
	}
	return r
}

func (this *Organization) FillFromLedgerValue(bytes []byte) error {
	if err := json.Unmarshal(bytes, &this); err != nil {
		return err
	} else {
		return nil
	}
}
