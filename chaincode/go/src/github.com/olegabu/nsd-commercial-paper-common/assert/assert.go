package assert

import (
	"testing"
	"fmt"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func Equal(t *testing.T, actual string, expected string, message string ) {
	if actual != expected {
		fmt.Println("Assert error:", message, "Current value: ", actual," , Expected value: ", expected)
		t.FailNow()
	}
}


func Ok(t *testing.T, invokeResult pb.Response, message string ) {
	if invokeResult.Status != 200 {
		fmt.Println("Assert error:", message, "Operation failed with status: ", invokeResult.Status, invokeResult.Message)
		t.FailNow()
	}
}


