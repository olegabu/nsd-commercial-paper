package testutils

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math/big"
	"fmt"
	"crypto/x509"
	"crypto/x509/pkix"
	"time"
	"crypto/rsa"
	"encoding/pem"
	"crypto/rand"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type TestStub struct {
	*shim.MockStub

	args [][]byte

	cc shim.Chaincode

	caller string

	mainOrg string
}

func (stub *TestStub) GetArgs() [][]byte {
	return stub.args
}

func (stub *TestStub) GetStringArgs() []string {
	args := stub.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

func (stub *TestStub) GetFunctionAndParameters() (function string, params []string) {
	allargs := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}

func (stub *TestStub) SetCaller(org string) {
	stub.caller = org
}

func (stub *TestStub) SetMainOrganization(name string) {
	stub.mainOrg = name
}

// Implemented to have a possibility to test privileges
func (ts *TestStub) GetCreator() ([]byte, error) {
	org := ts.caller

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Println("Failed to generate serial number: %s", err)
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		Issuer: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate private key: %s", err)
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s", err)
		return nil, err
	}

	result := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	return result, nil
}

// Reimplemented to have a possibility to test privileges
// NOTE: you should set caller and mainOrg to emulate call of book cc from depositary channel
func (stub *TestStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	if chaincodeName == "book" && channel == "depository" {
		if stub.caller != stub.mainOrg {
			return pb.Response{Status: 403, Message: "Insufficient privileges."}
		}
		return shim.Success([]byte(stub.mainOrg))
	}

	return shim.Success(nil)
}

// Initialise this chaincode,  also starts and ends a transaction.
// NOTE: you should set caller (if it matters) before using this function
func (stub *TestStub) MockInit(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

// Invoke this chaincode, also starts and ends a transaction.
// NOTE: you should set caller (if it matters) before using this function
func (stub *TestStub) MockInvoke(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

func NewTestStub(name string, cc shim.Chaincode) *TestStub {
	ts := &TestStub{MockStub: shim.NewMockStub(name, cc), cc: cc}
	return ts
}