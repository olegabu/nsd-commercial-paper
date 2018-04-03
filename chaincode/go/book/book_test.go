package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/Altoros/nsd-commercial-paper/chaincode/go/security"
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

	checkInit(t, stub, [][]byte{[]byte("init"), []byte("{\"mainOrg\":\"nsd.nsd.ru\", \"initEntries\":[{\"account\":\"AC0689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0001\",\"quantity\":\"100\"},{\"account\":\"AC0689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0002\",\"quantity\":\"42\"}]}")})

	//Correct transaction
	checkState(t, stub, 200, [][]byte{[]byte("check"), []byte("AC0689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("90")})
	//Wrong number of arguments
	checkState(t, stub, 400, [][]byte{[]byte("check"), []byte("AC0689654902")})
	// Record not found
	checkState(t, stub, 404, [][]byte{[]byte("check"), []byte("AAA"), []byte("BBB"), []byte("CCC"), []byte("200")})
	// Quantity less than current balance
	checkState(t, stub, 409, [][]byte{[]byte("check"), []byte("AC0689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("200")})
}

//TODO: uncomment when package for security changed to  "security"
//func TestRedeem(t *testing.T) {
//	sccSecurity := new(security.SecurityChaincode)
//	stubSecurity := shim.NewMockStub("security", sccSecurity)
//	stubSecurity.MockInit("1", [][]byte{[]byte("init"), []byte("RU000ABC0001"), []byte("active"), []byte("AAA689654902"), []byte("87680000045800005")})
//
//	fmt.Println(stubSecurity.State)
//
//	sccBook := new(BookChaincode)
//	stub := shim.NewMockStub("book", sccBook)
//
//	stub.MockPeerChaincode("security/common", stubSecurity)
//	checkInit(t, stub, [][]byte{[]byte("init"), []byte("[{\"account\":\"BBB689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0001\",\"quantity\":\"100\"}]")})
//
//	stub.MockInvoke("1", [][]byte{[]byte("redeem"), []byte("RU000ABC0001"), []byte("some message")})
//
//	//AAA should have at least 100
//	checkState(t, stub, 200, [][]byte{[]byte("check"), []byte("AAA689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("90")})
//	//BBB should have nothing on it's balance
//	checkState(t, stub, 409, [][]byte{[]byte("check"), []byte("BBB689654902"), []byte("87680000045800005"), []byte("RU000ABC0001"), []byte("90")})
//	//Second redeem is impossible
//	checkState(t, stub, 400, [][]byte{[]byte("redeem"), []byte("RU000ABC0001")})
//}