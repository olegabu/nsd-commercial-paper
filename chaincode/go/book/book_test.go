package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, expectedStatus int32,  args [][]byte) {
	bytes := stub.MockInvoke("1", args)
	if bytes.Status != expectedStatus {
		fmt.Println("Wrong status. Current value: ", bytes.Status,", Expected value: ", expectedStatus, ".")
		t.FailNow()
	}
}

func TestBook_Init(t *testing.T) {
	scc := new(BookChaincode)
	stub := shim.NewMockStub("bookChaincode", scc)

	checkInit(t, stub, [][]byte{[]byte("init"), []byte("AC0689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("100")})

	//Correct transaction
	checkState(t, stub, 200, [][]byte{[]byte("check"), []byte("AC0689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("90")})
	//Wrong number of arguments
	checkState(t, stub, 400, [][]byte{[]byte("check"), []byte("AC0689654902")})
	// Record not found
	checkState(t, stub, 404, [][]byte{[]byte("check"), []byte("AAA"), []byte("BBB"), []byte("CCC"), []byte("200")})
	// Quantity less than current balance
	checkState(t, stub, 409, [][]byte{[]byte("check"), []byte("AC0689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("200")})
}