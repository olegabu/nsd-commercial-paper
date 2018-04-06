package main
//package security

import (
	"fmt"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/Altoros/nsd-commercial-paper-common"
	"github.com/Altoros/nsd-commercial-paper-common/certificates"
)

var logger = shim.NewLogger("SecurityChaincode")

const indexName = `Security`

const EntryMaturedStatus = `MCAL`
const SecurityMaturedStatus = `matured`

// SecurityChaincode
type SecurityChaincode struct {
}

type SecurityValue struct {
	Status      	string 				`json:"status"`
	Entries			[]CalendarEntries	`json:"entries"`
	Redeem			nsd.Balance			`json:"redeem"`
}

type Security struct {
	Security        string 				`json:"security"`
	Status      	string 				`json:"status"`
	Entries			[]CalendarEntries	`json:"entries"`
	Redeem			nsd.Balance			`json:"redeem"`
}

type CalendarEntries struct {
	Date			string	`json:"date"`
	Code 			string	`json:"code"`
	Text			string	`json:"text"`
	Reference		string	`json:"reference"`
}

type KeyModificationValue struct {
	TxId      string 			`json:"txId"`
	Value     SecurityValue  	`json:"value"`
	Timestamp string 			`json:"timestamp"`
	IsDelete  bool   			`json:"isDelete"`
}

func (t *SecurityChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### SecurityChaincode Init ###########")

	_, args := stub.GetFunctionAndParameters()

	var securities []Security
	if err := json.Unmarshal([]byte(args[0]), &securities); err == nil && len(securities) != 0 {
		for _, entry := range securities {
			if rs := t.put(stub, []string{entry.Security, entry.Status, entry.Redeem.Account, entry.Redeem.Division});
			   rs.Status >= 400 {
				return rs
			}
		}
	} else {
		return pb.Response{Status: 400, Message: "JSON unmarshalling error."}
	}

	return shim.Success(nil)
}

func (t *SecurityChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### SecurityChaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "put" {
		return t.put(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}
	if function == "history" {
		return t.history(stub, args)
	}
	if function == "addEntry" {
		return t.addCalendarEntry(stub, args)
	}
	if function == "find" {
		return t.find(stub, args)
	}

	return shim.Error(fmt.Sprintf("Unknown function, check the first argument, must be one of: " +
		"put, query, history, addEntry, find. But got: %v", function))
}

func (t *SecurityChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	rs := stub.InvokeChaincode("book", [][]byte{[]byte("mainOrg")}, "depository")
	if rs.Status >= 400 {
		return pb.Response{Status: 400, Message: "Unable to invoke \"book\": " + rs.Message}
	}

	mainOrg := string(rs.Payload)
	if certificates.GetCreatorOrganization(stub) != mainOrg {
		return pb.Response{Status: 403,
			Message: "Insufficient privileges. Only " + mainOrg + " can change Securities information."}
	}

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security, status, Redeem Account, Redeem Division")
	}

	s, err := t.findByKey(stub, args[0])
	if err != nil {
		s.Security = args[0]
		s.Entries = []CalendarEntries{}
	}

	s.Status = args[1]
	s.Redeem = nsd.Balance{}
	s.Redeem.Account = args[2]
	s.Redeem.Division = args[3]

	return t.save(stub, s)
}

func (t *SecurityChaincode) save(stub shim.ChaincodeStubInterface, item Security) pb.Response {
	key, err := stub.CreateCompositeKey(indexName, []string{item.Security})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(SecurityValue{Status: item.Status,
											Entries: item.Entries,
											Redeem:nsd.Balance{Account:item.Redeem.Account, Division:item.Redeem.Division}})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SecurityChaincode) addCalendarEntry(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	rs := stub.InvokeChaincode("book", [][]byte{[]byte("mainOrg")}, "depository")
	if rs.Status >= 400 {
		return pb.Response{Status: 400, Message: "Unable to invoke \"book\": " + rs.Message}
	}

	mainOrg := string(rs.Payload)
	if certificates.GetCreatorOrganization(stub) != mainOrg {
		return pb.Response{Status: 403,
			Message: "Insufficient privileges. Only " + mainOrg + " can add Calendar Entry."}
	}

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security, code, date, text, reference")
	}

	security, err := t.findByKey(stub, args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Security not found: %v ", err))
	}

	entry := CalendarEntries{}
	entry.Code 		= args[1]
	entry.Date 		= args[2]
	entry.Text 		= args[3]
	entry.Reference = args[4]

	security.Entries = append(security.Entries, entry)

	if entry.Code == EntryMaturedStatus{
		security.Status = SecurityMaturedStatus
	}

	t.save(stub, security)

	return shim.Success(nil)
}

func (t *SecurityChaincode) findByKey(stub shim.ChaincodeStubInterface, securityName string) (Security, error) {

	key, err := stub.CreateCompositeKey(indexName, []string{securityName})
	if err != nil {
		return Security{}, fmt.Errorf("Cannot create composite key: %v", err)
	}

	response, err := stub.GetState(key)
	if err != nil {
		return Security{}, fmt.Errorf("Cannot read the state: %v", err)
	}
	if response == nil {
		return Security{}, fmt.Errorf("No security found for key: %v", key)
	}
	var value SecurityValue
	err = json.Unmarshal(response, &value)
	if err != nil {
		return Security{}, fmt.Errorf("Cannot Unmarshal security: %v", err)
	}

	security := Security {
		Security: securityName,
		Status: value.Status,
		Redeem: nsd.Balance{
			Account: value.Redeem.Account,
			Division: value.Redeem.Division,
		},
		Entries:value.Entries,
	}

	return security, nil
}

func (t *SecurityChaincode) find(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security")
	}

	security, err := t.findByKey(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	result, err := json.Marshal(security)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *SecurityChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	securities := []Security{}
	for it.HasNext() {
		responseRange, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		//account-division-security
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		var value SecurityValue
		err = json.Unmarshal(responseRange.Value, &value)
		if err != nil {
			return shim.Error(err.Error())
		}

		security := Security {
			Security: compositeKeyParts[0],
			Status: value.Status,
			Redeem: nsd.Balance{
				Account:  value.Redeem.Account,
				Division: value.Redeem.Division,
			},
			Entries:value.Entries,
		}

		securities = append(securities, security)
	}

	result, err := json.Marshal(securities)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *SecurityChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. " +
			"Expecting security")
	}

	//account-division-security
	key, err := stub.CreateCompositeKey(indexName, args)
	if err != nil {
		return shim.Error(err.Error())
	}

	it, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	modifications := []KeyModificationValue{}

	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var entry KeyModificationValue

		entry.TxId = response.GetTxId()
		entry.IsDelete = response.GetIsDelete()
		ts := response.GetTimestamp()

		if ts != nil {
			entry.Timestamp = time.Unix(ts.Seconds, int64(ts.Nanos)).String()
		}

		err = json.Unmarshal(response.GetValue(), &entry.Value)
		if err != nil {
			return shim.Error(err.Error())
		}

		modifications = append(modifications, entry)
	}

	result, err := json.Marshal(modifications)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(SecurityChaincode))
	if err != nil {
		logger.Errorf("Error starting Security chaincode: %s", err)
	}
}

