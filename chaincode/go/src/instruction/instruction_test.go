package main

import (
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/olegabu/nsd-commercial-paper-common/assert"
)



func getInvokeArgs(command string, args []string) [][]byte {
	byteArgs := [][]byte{}
	byteArgs = append(byteArgs, []byte(command))
	for _, arg := range args {
		byteArgs = append(byteArgs, []byte(arg))
	}
	return byteArgs
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
	res := stub.MockInit("123", initArgs)

	assert.Ok(t, res, "Init failed");
	return stub
}

func Test_InstructionInit(t *testing.T) {
	stub := initInstructionCC(t)

	// GetCreator is not implemented in NewMockStub
	//stub.GetCreator()

	key,  _ := stub.CreateCompositeKey("Authentication", []string{"RBIOWNER0ACC", "00000000000000000"})
	data, _ := stub.GetState(key)


	org2 := "{\"organization\":\"raiffeisen.nsd.ru\"," +
		"\"deponent\":\"DE000DB7HWY7\"," +
		"\"balances\":[{\"account\":\"RBIOWNER0ACC\",\"division\":\"00000000000000000\"}]}"
	assert.Equal(t, string(data), org2, "Initialize instruction data");
}



func Test_TransferFOP_legacy(t *testing.T) {
	//stub := shim.NewMockStub("instruction", new(InstructionChaincode))
	stub := initInstructionCC(t)


	transferArguments := getInvokeArgs("transfer", []string{
		//"fop", No type here!
		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",        		 // quantity
		"ref-123",           // reference
		"2017-12-31",        // instructionDate
		"2017-12-31",        // tradeDate

		"DE000DB7HWY7", 	// deponentFrom
		"CA9861913023", 	// deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason
	})


	res := stub.MockInvoke("1", transferArguments)
	assert.Ok(t, res, "Transfer failed");

}



func Test_TransferFOP(t *testing.T) {
	//stub := shim.NewMockStub("instruction", new(InstructionChaincode))
	stub := initInstructionCC(t)

	transferArguments := getInvokeArgs("transfer", []string{
		"fop",
		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",        		 // quantity
		"ref-123",           // reference
		"2017-12-31",        // instructionDate
		"2017-12-31",        // tradeDate

		"DE000DB7HWY7", 	// deponentFrom
		"CA9861913023", 	// deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason
	})

	res := stub.MockInvoke("1", transferArguments)
	assert.Ok(t, res, "Transfer failed");

	//
	//

	transfer2 := getInvokeArgs("transfer", []string{
		"fop",
		"MZ130605006C",
		"19000000000000000",
		"MS980129006C",
		"00000000000000000",
		"RU000A0JWGG3",
		"1",
		"test1",
		"2018-03-04",
		"2018-03-04",
		"CA9861913023",
		"DE000DB7HWY7",
		"123",
		"{}",
	})

	res2 := stub.MockInvoke("2", transfer2)
	assert.Ok(t, res2, "Transfer2 failed");

}



func Test_TransferDVP(t *testing.T) {
	//stub := shim.NewMockStub("instruction", new(InstructionChaincode))
	stub := initInstructionCC(t)


	transferArguments := getInvokeArgs("transfer", []string{
		"dvp", // type

		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",         		 // quantity
		"ref-123",           // reference
		"2017-12-31",        // instructionDate
		"2017-12-31",        // tradeDate

		"DE000DB7HWY7", 	 // deponentFrom
		"CA9861913023", 	 // deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason



		"40701810000000001000", // transfererAccount
		"f044525505op",     	// transfererBic

		"40701810000000001000", // receiverAccount
		"f044525505op",     	// receiverBic

		"30000000",       		// paymentAmount
		"RUB",          		// paymentCurrency
	})


	res := stub.MockInvoke("1", transferArguments)
	assert.Ok(t, res, "Transfer failed");
}

