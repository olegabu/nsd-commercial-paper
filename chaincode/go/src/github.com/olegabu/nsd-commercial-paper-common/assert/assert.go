package assert

import (
	"testing"
	"fmt"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

func NotError(t *testing.T, actual error) {
	if actual != nil {
		fmt.Println("Assert error:", actual.Error())
		t.FailNow()
	}
}


func Equal(t *testing.T, actual string, expected string, message ...string ) {
	if actual != expected {

		errorMessage := "Equal fail."
		if len(message) > 0 {
			errorMessage = message[0]
		}

		fmt.Println("Assert error:", errorMessage, "Current value: ", actual," , Expected value: ", expected)
		t.FailNow()
	}
}


func Ok(t *testing.T, invokeResult pb.Response, message ...string) {
	if invokeResult.Status != 200 {

		errorMessage := "Ok fail."
		if len(message) > 0 {
			errorMessage = message[0]
		}

		fmt.Println("Assert error:", errorMessage, "Operation failed with status: ", invokeResult.Status, invokeResult.Message)
		t.FailNow()
	}
}


func IsNumber(message string ) bool {
	if _, err := strconv.Atoi(message); err != nil {
		return false
	}
	return true
}

