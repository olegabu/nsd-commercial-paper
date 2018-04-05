package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"github.com/Altoros/nsd-commercial-paper-common/testutils"
)

const nsdName = "nsd.nsd.ru"

func getStub(t *testing.T) *testutils.TestStub{
	scc := new(SecurityChaincode)
	ts := testutils.NewTestStub("security", scc)
	ts.SetCaller(nsdName)
	ts.SetMainOrganization(nsdName)
	return ts
}
func getInitializedStub(t *testing.T) *testutils.TestStub{
	stub := getStub(t)
	stub.MockInit("1", [][]byte{[]byte("init"), []byte(
		`[{
			"security":"RU000ABC0001",
			"status":"active",
			"redeem":{
				"account":"AC0689654902",
				"division":"87680000045800005"
			}
		}]`)})
	return stub
}

func checkInit(t *testing.T, stub *testutils.TestStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed: ", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *testutils.TestStub, expectedStatus int32,  args [][]byte) []Security {
	bytes := stub.MockInvoke("1", args)
	if bytes.Status != expectedStatus {
		fmt.Println("Wrong status. Current value: ", bytes.Status,", Expected value: ", expectedStatus, ".")
		t.FailNow()
	}
	var value []Security
	err := json.Unmarshal(bytes.Payload, &value)
	if err != nil {
		fmt.Println("Cannot Unmarshal security: %v", err)
		t.FailNow()
	}
	return value
}

func TestSecurity_Init(t *testing.T) {
	checkInit(t, getStub(t), [][]byte{[]byte("init"), []byte(
		`[{
			"security":"RU000ABC0001",
			"status":"active",
			"redeem":{
				"account":"AC0689654902",
				"division":"87680000045800005"
			}
		}]`)})
}

func TestSecurity_Query(t *testing.T) {
	stub := getInitializedStub(t)
	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})
	securityName := "RU000ABC0001"
	securityStatus:= "active"
	redeemAccount := "AC0689654902"
	redeemDivision:= "87680000045800005"

	if len(securities) != 1 {
		fmt.Println("Security not found")
		t.FailNow()
	}
	if securities[0].Security != securityName {
		fmt.Println("Newly created security has wrong name :", securities[0].Security, " , expected: ", securityName)
		t.FailNow()
	}
	if securities[0].Status != securityStatus {
		fmt.Println("Newly created security has wrong status :", securities[0].Status, " , expected: ", securityStatus)
		t.FailNow()
	}
	if securities[0].Redeem.Account !=  redeemAccount{
		fmt.Println("Newly created security has wrong RedeemAccount :", securities[0].Redeem.Account, " , expected: ", redeemAccount)
		t.FailNow()
	}
	if securities[0].Redeem.Division != redeemDivision {
		fmt.Println("Newly created security has wrong RedeemDivision :", securities[0].Redeem.Division, " , expected: ", redeemDivision)
		t.FailNow()
	}
	if len(securities[0].Entries) != 0 {
		fmt.Println("Newly created security has calendar entries. Expected empty list")
		t.FailNow()
	}

}

func TestSecurity_Put(t *testing.T){
	stub := getStub(t)

	securityName := "RU000ABC0002"
	securityStatus:= "created"
	redeemAccount := "AC0689654902"
	redeemDivision:= "87680000045800005"

	stub.MockInvoke("1", [][]byte{[]byte("put"), []byte(securityName), []byte(securityStatus), []byte(redeemAccount), []byte(redeemDivision)})

	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})

	if len(securities) != 1{
		fmt.Println("Security was not created correctly.")
		t.FailNow()
	}
	if securities[0].Security != securityName {
		fmt.Println("Newly created security has wrong name :", securities[0].Security, " , expected: ", securityName)
		t.FailNow()
	}
	if securities[0].Status != securityStatus {
		fmt.Println("Newly created security has wrong status :", securities[0].Status, " , expected: ", securityStatus)
		t.FailNow()
	}
	if securities[0].Redeem.Account !=  redeemAccount{
		fmt.Println("Newly created security has wrong RedeemAccount :", securities[0].Redeem.Account, " , expected: ", redeemAccount)
		t.FailNow()
	}
	if securities[0].Redeem.Division != redeemDivision {
		fmt.Println("Newly created security has wrong RedeemDivision :", securities[0].Redeem.Division, " , expected: ", redeemDivision)
		t.FailNow()
	}
	if len(securities[0].Entries) != 0 {
		fmt.Println("Newly created security has calendar entries. Expected empty list")
		t.FailNow()
	}
}

func TestSecurity_AddEntry(t *testing.T){
	stub := getInitializedStub(t)

	name := "RU000ABC0001"
	code := "updated"
	date := "12/12/17"
	text := "Some message"
	reference := "#35"

	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(name), []byte(code), []byte(date), []byte(text), []byte(reference)})

	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})

	if len(securities) != 1{
		fmt.Println("Security was not created correctly.")
		t.FailNow()
	}
	if len(securities[0].Entries) != 1 {
		fmt.Println("New Calendar Entry was not created")
		t.FailNow()
	}
	if securities[0].Entries[0].Code != code {
		fmt.Println("Newly created Entry has wrong code :", securities[0].Entries[0].Code, " , expected: ", code)
		t.FailNow()
	}
	if securities[0].Entries[0].Date != date {
		fmt.Println("Newly created Entry has wrong date :", securities[0].Entries[0].Date, " , expected: ", date)
		t.FailNow()
	}
	if securities[0].Entries[0].Text != text {
		fmt.Println("Newly created Entry has wrong text :", securities[0].Entries[0].Text, " , expected: ", text)
		t.FailNow()
	}
	if securities[0].Entries[0].Reference != reference {
		fmt.Println("Newly created Entry has wrong reference :", securities[0].Entries[0].Reference, " , expected: ", reference)
		t.FailNow()
	}
}

func TestSecurity_AddMultipleEntries(t *testing.T){
	stub := getInitializedStub(t)

	name := "RU000ABC0001"
	code := "updated"
	date := "12/12/17"
	text := "Some message"
	reference := "#35"

	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(name), []byte(code), []byte(date), []byte(text), []byte(reference)})
	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(name), []byte(code), []byte(date), []byte(text), []byte(reference)})
	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(name), []byte(code), []byte(date), []byte(text), []byte(reference)})

	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})

	if len(securities) != 1{
		fmt.Println("Security was not created correctly.")
		t.FailNow()
	}
	if len(securities[0].Entries) != 3 {
		fmt.Println("One of Calendar Entries was not created")
		t.FailNow()
	}
}

func TestSecurity_EntryMaturity(t *testing.T){
	stub := getInitializedStub(t)

	name := "RU000ABC0001"
	code := EntryMaturedStatus
	date := "12/12/17"
	text := "Some message"
	reference := "#35"

	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(name), []byte(code), []byte(date), []byte(text), []byte(reference)})

	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})

	if len(securities) != 1{
		fmt.Println("Security was not created correctly.")
		t.FailNow()
	}
	if securities[0].Status != SecurityMaturedStatus{
		fmt.Println("Security state should be changed to: ", SecurityMaturedStatus, ", uppon receiving Entry with code: ", EntryMaturedStatus)
		t.FailNow()
	}
}

func TestSecurity_Update(t *testing.T){
	stub := getInitializedStub(t)

	securityName := "RU000ABC0001"
	newStatus:= "created"
	redeemAccount := "AC0689654902"
	redeemDivision:= "87680000045800005"

	code := "updated"
	date := "12/12/17"
	text := "Some message"
	reference := "#35"

	stub.MockInvoke("1", [][]byte{[]byte("addEntry"), []byte(securityName), []byte(code), []byte(date), []byte(text), []byte(reference)})


	stub.MockInvoke("1", [][]byte{[]byte("put"), []byte(securityName), []byte(newStatus), []byte(redeemAccount), []byte(redeemDivision)})

	securities := checkState(t, stub, 200, [][]byte{[]byte("query")})

	if len(securities) != 1{
		fmt.Println("Security was not created correctly.")
		t.FailNow()
	}

	if securities[0].Security != securityName {
		fmt.Println("Newly created security has wrong name")
		t.FailNow()
	}
	if securities[0].Status != newStatus {
		fmt.Println("Newly created security has wrong status")
		t.FailNow()
	}
	if securities[0].Redeem.Account !=  redeemAccount{
		fmt.Println("Newly created security has wrong RedeemAccount :", securities[0].Redeem.Account, " , expected: ", redeemAccount)
		t.FailNow()
	}
	if securities[0].Redeem.Division != redeemDivision {
		fmt.Println("Newly created security has wrong RedeemDivision :", securities[0].Redeem.Division, " , expected: ", redeemDivision)
		t.FailNow()
	}
	if len(securities[0].Entries) != 1 {
		fmt.Println("Previously created Entry was deleted during update")
		t.FailNow()
	}

}