package nsd


import (
	"testing"
	"github.com/olegabu/nsd-commercial-paper-common/assert"
)



func Test_FillFromCompositeKeyParts_FOP(t *testing.T) {
	instruction := Instruction{}

	transferArguments := []string{
		"fop", // type

		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",         		 // quantity
		"ref-23",            // reference
		"2017-2-31",         // instructionDate
		"2017-2-31",         // tradeDate

		"DE000DB7HWY7", // deponentFrom
		"CA9861913023", // deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason
	}

	err := instruction.FillFromArgs(transferArguments, "transferer")
	assert.NotError(t, err)

	assert.Equal(t, string(instruction.Key.Type), "fop")
	assert.Equal(t, string(instruction.Key.Transferer.Account), "MFONISSUEACC")
	assert.Equal(t, string(instruction.Key.Transferer.Division), "19000000000000000")
	assert.Equal(t, string(instruction.Key.Receiver.Account), "RBIOWNER0ACC")
	assert.Equal(t, string(instruction.Key.Receiver.Division), "00000000000000000")
	assert.Equal(t, string(instruction.Key.Security), "RU000ABC0001")

}

func Test_FillFromCompositeKeyParts_FOP_legacy(t *testing.T) {
	instruction := Instruction{}

	transferArguments := []string{
		//"fop", // type

		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",         		 // quantity
		"ref-23",            // reference
		"2017-2-31",         // instructionDate
		"2017-2-31",         // tradeDate

		"DE000DB7HWY7", // deponentFrom
		"CA9861913023", // deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason
	}

	err := instruction.FillFromArgs(transferArguments, "transferer")
	assert.NotError(t, err)

	assert.Equal(t, string(instruction.Key.Type), "fop")
	assert.Equal(t, string(instruction.Key.Transferer.Account), "MFONISSUEACC")
	assert.Equal(t, string(instruction.Key.Transferer.Division), "19000000000000000")
	assert.Equal(t, string(instruction.Key.Receiver.Account), "RBIOWNER0ACC")
	assert.Equal(t, string(instruction.Key.Receiver.Division), "00000000000000000")
	assert.Equal(t, string(instruction.Key.Security), "RU000ABC0001")

}



func Test_FillFromCompositeKeyParts_DVP(t *testing.T) {
	instruction := Instruction{}

	transferArguments := []string {
		"dvp", // type

		"MFONISSUEACC",      // accountFrom
		"19000000000000000", // divisionFrom

		"RBIOWNER0ACC",      // accountTo
		"00000000000000000", // divisionTo

		"RU000ABC0001",      // security
		"123",         		 // quantity
		"ref-23",            // reference
		"2017-2-31",         // instructionDate
		"2017-2-31",         // tradeDate

		"40701810000000001000", // transfererAccount
		"f044525505op",     // transfererBic

		"40701810000000001001", // receiverAccount
		"f044525505oq",     // receiverBic

		"30000000",       // paymentAmount
		"RUB",          // paymentCurrency


		"DE000DB7HWY7", // deponentFrom
		"CA9861913023", // deponentTo
		"memberInstructionId", // memberInstructionId
		"{\"json_reason\":\"any json\"}", // reason
	}

	err := instruction.FillFromArgs(transferArguments, "transferer")
	assert.NotError(t, err)

	assert.Equal(t, string(instruction.Key.Type), "dvp")
	assert.Equal(t, string(instruction.Key.Transferer.Account), "MFONISSUEACC")
	assert.Equal(t, string(instruction.Key.Transferer.Division), "19000000000000000")
	assert.Equal(t, string(instruction.Key.Receiver.Account), "RBIOWNER0ACC")
	assert.Equal(t, string(instruction.Key.Receiver.Division), "00000000000000000")
	assert.Equal(t, string(instruction.Key.Security), "RU000ABC0001")

	assert.Equal(t, string(instruction.Key.PaymentFrom.Account), "40701810000000001000")
	assert.Equal(t, string(instruction.Key.PaymentFrom.Bic), "f044525505op")
	assert.Equal(t, string(instruction.Key.PaymentTo.Account), "40701810000000001001")
	assert.Equal(t, string(instruction.Key.PaymentTo.Bic), "f044525505oq")
	assert.Equal(t, string(instruction.Key.PaymentAmount), "30000000")
	assert.Equal(t, string(instruction.Key.PaymentCurrency), "RUB")

}


