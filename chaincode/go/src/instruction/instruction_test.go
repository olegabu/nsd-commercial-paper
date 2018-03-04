package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/olegabu/nsd-commercial-paper-common/assert"
)


func checkState(t *testing.T, stub *shim.MockStub, expectedStatus int32,  args []string) {

	byteArgs := [][]byte{}
	for _, arg := range args {
		byteArgs = append(byteArgs, []byte(arg))
	}

	bytes := stub.MockInvoke("1", byteArgs)
	if bytes.Status != expectedStatus {
		fmt.Println("Wrong status. Current value: ", bytes.Status,", Expected value: ", expectedStatus, ".")
		t.FailNow()
	}
}


func initInstructionCC(t *testing.T) *shim.MockStub {

	stub := shim.NewMockStub("instruction", new(InstructionChaincode))

	org1 := "{" +
		"\"organization\":\"megafon.nsd.ru\"," +
		"\"deponent\":\"CA9861913023\"," +
		"\"balances\":[{\"account\":\"MFONISSUEACC\",\"division\":\"19000000000000000\"}," +
		"{\"account\":\"MFONISSUEACC\",\"division\":\"22000000000000000\"}]}"

	org2 := "{\"organization\":\"raiffeisen.nsd.ru\"," +
		"\"deponent\":\"DE000DB7HWY7\"," +
		"\"balances\":[{\"account\":\"RBIOWNER0ACC\",\"division\":\"00000000000000000\"}]}"

	//
	initArgs := [][]byte{
		[]byte("init"),

		[]byte("[" + org1 + "," + org2 + "]"),
	}
	res := stub.MockInit("1", initArgs)

	assert.Ok(t, res, "Init failed");
	return stub
}

func Test_InstructionInit(t *testing.T) {
	stub := initInstructionCC(t)

	// GetCreator is not implemented in NewMockStub
	//stub.GetCreator()


	key,  _ := stub.CreateCompositeKey(authenticationIndex, []string{"RBIOWNER0ACC", "00000000000000000"})
	data, _ := stub.GetState(key)


	org2 := "{\"organization\":\"raiffeisen.nsd.ru\"," +
		"\"deponent\":\"DE000DB7HWY7\"," +
		"\"balances\":[{\"account\":\"RBIOWNER0ACC\",\"division\":\"00000000000000000\"}]}"
	assert.Equal(t, string(data), org2, "Initialize instruction data");
}



func Test_TransferFOP(t *testing.T) {
	//stub := shim.NewMockStub("instruction", new(InstructionChaincode))
	stub := initInstructionCC(t)


	transferArguments := [][]byte{
		[]byte("transfer"),

		[]byte("MFONISSUEACC"),      // accountFrom
		[]byte("19000000000000000"), // divisionFrom

		[]byte("RBIOWNER0ACC"),      // accountTo
		[]byte("00000000000000000"), // divisionTo

		[]byte("RU000ABC0001"),      // security
		[]byte("123"),				 // quantity
		[]byte("ref-123"),           // reference
		[]byte("2017-12-31"),        // instructionDate
		[]byte("2017-12-31"),        // tradeDate

		[]byte("DE000DB7HWY7"), // deponentFrom
		[]byte("CA9861913023"), // deponentTo
		[]byte("memberInstructionId"), // memberInstructionId
		[]byte("{\"json_reason\":\"any json\"}"), // reason
	}


	res := stub.MockInvoke("1", transferArguments)
	assert.Ok(t, res, "Transfer failed");
}



func Test_TransferDVP(t *testing.T) {
	//stub := shim.NewMockStub("instruction", new(InstructionChaincode))
	stub := initInstructionCC(t)


	transferArguments := [][]byte{
		[]byte("transfer"),

		[]byte("MFONISSUEACC"),      // accountFrom
		[]byte("19000000000000000"), // divisionFrom

		[]byte("RBIOWNER0ACC"),      // accountTo
		[]byte("00000000000000000"), // divisionTo

		[]byte("RU000ABC0001"),      // security
		[]byte("123"),				 // quantity
		[]byte("ref-123"),           // reference
		[]byte("2017-12-31"),        // instructionDate
		[]byte("2017-12-31"),        // tradeDate

		[]byte("DE000DB7HWY7"), // deponentFrom
		[]byte("CA9861913023"), // deponentTo
		[]byte("memberInstructionId"), // memberInstructionId
		[]byte("{\"json_reason\":\"any json\"}"), // reason


		[]byte("dvp"), // type

		[]byte("40701810000000001000"), // transfererAccount
		[]byte("f044525505op"), 		// transfererBic

		[]byte("40701810000000001000"), // receiverAccount
		[]byte("f044525505op"), 		// receiverBic

		[]byte("30000000"), 			// paymentAmount
		[]byte("RUB"), 					// paymentCurrency

	}


	res := stub.MockInvoke("1", transferArguments)
	assert.Ok(t, res, "Transfer failed");
}

