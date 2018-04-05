package certificates

import (
	"strings"
	"encoding/pem"
	"crypto/x509"
	"io/ioutil"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("Certificates Common")

// **** Security Methods **** //
func getOrganization(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	return organization
}

func GetCreatorOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	return getOrganization(certificate)
}

func GetMyOrganization() string {
	// TODO get the filename from $CORE_PEER_TLS_ROOTCERT_FILE
	// better way perhaps is to pass a flag in transient map to nsd peer to ask to check against book chaincode
	certFilename := "/etc/hyperledger/fabric/peer.crt"
	certificate, err := ioutil.ReadFile(certFilename)
	if err != nil {
		logger.Debugf("cannot read my peer's certificate file %s", certFilename)
		return ""
	}

	return getOrganization(certificate)
}
