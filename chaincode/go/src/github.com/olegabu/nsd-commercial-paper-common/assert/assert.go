package assert

import (
	"testing"
	"fmt"
)

func Equal(t *testing.T, actual string, expected string, message string ) {
	if actual != expected {
		fmt.Println("Assert error:", message, "Current value: ", actual," , Expected value: ", expected)
		t.FailNow()
	}
}


