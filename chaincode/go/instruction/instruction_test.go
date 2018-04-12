package main

import (
	"testing"
	"github.com/Altoros/nsd-commercial-paper-common"
	"github.com/Altoros/nsd-commercial-paper-common/testutils"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"sort"
)

const nsdName = "nsd.nsd.ru"

type queryResult struct {
	Name    string      `json:"organization"`
	Balance nsd.Balance `json:"balance"`
}

func toByteArray(arr []string) [][]byte {
	var res [][]byte
	for _, entry := range(arr) {
		res = append(res, []byte(entry))
	}

	return res
}

func getStub(t *testing.T) *testutils.TestStub{
	cc := new(InstructionChaincode)
	return testutils.NewTestStub("instruction", cc)
}

func getInitializedStub(t *testing.T) *testutils.TestStub{
	stub := getStub(t)
	args := [][]byte{[]byte("init"), []byte(
		`[{
			"organization": "org1",
			"balances": [
				{
					"account": "MZ0987654321",
					"division": "19000000000000000"
				}
			]
		}, {
			"organization": "org2",
			"balances": [
				{
					"account": "30109810000000000000",
					"division": "044525505"
				}
			]
		}]`)}

	stub.SetCaller(nsdName)
	stub.SetMainOrganization(nsdName)
	stub.MockInit("1", args)

	return stub
}

func TestInstructionChaincode_Init(t *testing.T) {
	stub := getStub(t)
	args := [][]byte{[]byte("init"), []byte(
		`[{
			"organization": "org1",
			"balances": [
				{
					"account": "MZ0987654321",
					"division": "19000000000000000"
				}
			]
		}, {
			"organization": "org2",
			"balances": [
				{
					"account": "30109810000000000000",
					"division": "044525505"
				}
			]
		}]`)}

	stub.SetCaller(nsdName)
	res := stub.MockInit("1", args)

	if res.Status != shim.OK {
		fmt.Println("Init failed: ", string(res.Message))
		t.FailNow()
	}
}

func TestInstructionChaincode_ReceiveTransfer(t *testing.T) {
	stub := getInitializedStub(t)

	var response pb.Response

	// check invoking receive/transfer with insufficient number of arguments
	response = stub.MockInvoke("1", [][]byte{[]byte("receive")})
	if response.Status < 400 {
		fmt.Println(`"Receive" has succeeded with wrong number of arguments.`)
		t.FailNow()
	} else {
		fmt.Println("Receive predicted error: " + response.Message)
	}

	response = stub.MockInvoke("1", [][]byte{[]byte("transfer")})
	if response.Status < 400 {
		fmt.Println(`"Transfer" has succeeded with wrong number of arguments.`)
		t.FailNow()
	} else {
		fmt.Println("Transfer predicted error: " + response.Message)
	}

	//----- FOP Instruction Flow Testing -----//
	fmt.Println("//----- FOP Instruction Flow Testing -----//")

	baseRecvArgs := []string{"receive", "MZ0987654321", "19000000000000000", "30109810000000000000", "044525505",
	    "RU000A0JVVB5", "500", "SOMEREF123", "2018-03-29", "2018-03-29", "fop"}
	addRecvArgs := []string{"MCXXXXX00000", "MSYYYYY00000", "id_to",
		`{"document": "doc_to", "description": "321", "created": "2018-03-29"}`}

	baseTransfArgs := []string{"transfer", "MZ0987654321", "19000000000000000", "30109810000000000000", "044525505",
		"RU000A0JVVB5", "500", "SOMEREF123", "2018-03-29", "2018-03-29", "fop"}
	addTransfArgs := []string{"MCXXXXX00000", "MSYYYYY00000", "id_from",
		`{"document": "doc_from", "description": "123", "created": "2018-03-29"}`}

	// check if transferer can't call receive
	stub.SetCaller("org1")
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Receive" has been called by the transferer.`)
		t.FailNow()
	} else {
		fmt.Println("Receive predicted error: " + response.Message)
	}

	stub.SetCaller("org2")
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status >= 400 {
		fmt.Println("Receive error: " + response.Message)
		t.FailNow()
	}

	// check the same instruction processing
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Receive" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Receive predicted error: " + response.Message)
	}

	// check duplication processing
	baseRecvArgs[5], baseRecvArgs[6] = "RU111A1JVVB6", "1000"
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Receive" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Receive predicted error: " + response.Message)
	}

	// check if receiver can't call transfer
	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Transfer" has been called by the receiver.`)
		t.FailNow()
	} else {
		fmt.Println("Transfer predicted error: " + response.Message)
	}

	// check matching
	stub.SetCaller("org1")
	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status >= 400 {
		fmt.Println("Transfer error: " + response.Message)
		t.FailNow()
	}

	// check duplication processing
	baseTransfArgs[5], baseTransfArgs[6] = "RU111A1JVVB6", "1000"
	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Transfer" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Transfer predicted error: " + response.Message)
	}

	//----- DVP Instruction Flow Testing-----//
	fmt.Println("//----- DVP Instruction Flow Testing -----//")

	baseRecvArgs[len(baseRecvArgs) - 1] = "dvp"
	baseRecvArgs[7] = "ANOTHERREF123"
	baseRecvArgs = append(baseRecvArgs, "tr_money_acc", "tr_money_bic", "rc_money_acc", "rc_money_bic",
		"10000.00", "RUB")
	addRecvArgs = append(addRecvArgs, `{"description": "Additional info."}`)
	addRecvArgs[2] = "id_to_2"
	baseTransfArgs[len(baseTransfArgs) - 1] = "dvp"
	baseTransfArgs[7] = "ANOTHERREF123"
	baseTransfArgs = append(baseTransfArgs, "tr_money_acc", "tr_money_bic", "rc_money_acc", "rc_money_bic",
		"10000.00", "RUB")
	addTransfArgs[2] = "id_from_2"

	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status >= 400 {
		fmt.Println("Transfer error: " + response.Message)
		t.FailNow()
	}

	// check the same instruction processing
	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Transfer" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Transfer predicted error: " + response.Message)
	}

	// check duplication processing
	baseTransfArgs[5], baseTransfArgs[6] = "RU222A1JVVB7", "1500"
	response = stub.MockInvoke("1", toByteArray(append(baseTransfArgs, addTransfArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Transfer" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Transfer predicted error: " + response.Message)
	}

	// check matching
	stub.SetCaller("org2")
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status >= 400 {
		fmt.Println("Receive error: " + response.Message)
		t.FailNow()
	}

	// check duplication processing
	baseRecvArgs[5], baseRecvArgs[6] = "RU222A1JVVB7", "1500"
	response = stub.MockInvoke("1", toByteArray(append(baseRecvArgs, addRecvArgs...)))
	if response.Status < 400 {
		fmt.Println(`"Receive" has succeeded with duplicate.`)
		t.FailNow()
	} else {
		fmt.Println("Receive predicted error: " + response.Message)
	}
}

func checkBalanceQuery(results, expectedResults []queryResult) error {
	if len(results) != len(expectedResults) {
		return fmt.Errorf("Query result contains less elements then expected.")
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Balance.Account < results[j].Balance.Account
	})

	sort.Slice(expectedResults, func(i, j int) bool {
		return expectedResults[i].Balance.Account < expectedResults[j].Balance.Account
	})

	for i, _ := range results {
		if results[i].Name != expectedResults[i].Name ||
		   results[i].Balance.Account != expectedResults[i].Balance.Account ||
		   results[i].Balance.Division != expectedResults[i].Balance.Division {
			return fmt.Errorf("Query result #%d is not equal the expected one.", i)
		}
	}

	return nil
}

func TestInstructionChaincode_GetBalances(t *testing.T) {
	stub := getInitializedStub(t)

	response := stub.MockInvoke("1", [][]byte{[]byte("getBalances")})
	if response.Status >= 400 {
		fmt.Println(`"getBalances" error: ` + response.Message)
		t.FailNow()
	}

	var results []queryResult
	if err := json.Unmarshal(response.Payload, &results); err != nil {
		fmt.Println("JSON unmarshalling error: " + err.Error())
		t.FailNow()
	}

	expectedResults := []queryResult{
		queryResult{
			Name: "org1",
			Balance: nsd.Balance{
				Account: "MZ0987654321",
				Division: "19000000000000000",
			},
		},
		queryResult{
			Name: "org2",
			Balance: nsd.Balance{
				Account: "30109810000000000000",
				Division: "044525505",
			},
		},
	}

	if err := checkBalanceQuery(results, expectedResults); err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}

func TestInstructionChaincode_AddBalances(t *testing.T) {
	stub := getInitializedStub(t)

	expectedResults := []queryResult{
		queryResult{
			Name: "org1",
			Balance: nsd.Balance{
				Account: "MZ0987654321",
				Division: "19000000000000000",
			},
		},
		queryResult{
			Name: "org1",
			Balance: nsd.Balance{
				Account: "MZ0987654322",
				Division: "22000000000000000",
			},
		},
		queryResult{
			Name: "org2",
			Balance: nsd.Balance{
				Account: "30109810000000000000",
				Division: "044525505",
			},
		},
	}

	var response pb.Response

	// first time we're adding a new balance record; second time we're adding existing record
	// the behaviour isn't expected to change
	for i := 0; i < 2; i++ {
		response = stub.MockInvoke("1", [][]byte{[]byte("addBalances"), []byte(`[{
				"organization": "org1",
				"balances": [
					{
						"account": "MZ0987654322",
						"division": "22000000000000000"
					}
				]
			}]`)})
		if response.Status >= 400 {
			fmt.Println(`"addBalances" error: ` + response.Message)
			t.FailNow()
		}

		response = stub.MockInvoke("1", [][]byte{[]byte("getBalances")})
		if response.Status >= 400 {
			fmt.Println(`"getBalances" error: ` + response.Message)
			t.FailNow()
		}

		var results []queryResult
		if err := json.Unmarshal(response.Payload, &results); err != nil {
			fmt.Println("JSON unmarshalling error: " + err.Error())
			t.FailNow()
		}

		if err := checkBalanceQuery(results, expectedResults); err != nil {
			fmt.Println(err)
			t.FailNow()
		}
	}

	// trying to add balances without nsd permissions must lead to an error
	stub.SetCaller("org1")
	response = stub.MockInvoke("1", [][]byte{[]byte("addBalances"), []byte(`[{
			"organization": "org1",
			"balances": [
				{
					"account": "MZ0987654322",
					"division": "22000000000000000"
				}
			]
		}]`)})
	if response.Status < 400 {
		fmt.Println(`"addBalances" can be called without nsd permissions.`)
		t.FailNow()
	} else {
		fmt.Println(`Predicted "addBalances" error: ` + response.Message)
	}
}

func TestInstructionChaincode_RemoveBalances(t *testing.T) {
	stub := getInitializedStub(t)

	expectedResults := []queryResult{
		queryResult{
			Name: "org2",
			Balance: nsd.Balance{
				Account: "30109810000000000000",
				Division: "044525505",
			},
		},
	}

	var response pb.Response

	// first time we're removing an existing record; second time we're removing record that doesn't exist in ledger
	// the behaviour isn't expected to change
	for i := 0; i < 2; i++ {
		response = stub.MockInvoke("1", [][]byte{[]byte("removeBalances"), []byte(`[{
			"organization": "org1",
			"balances": [
				{
					"account": "MZ0987654321",
					"division": "19000000000000000"
				}
			]
		}]`)})
		if response.Status >= 400 {
			fmt.Println(`"removeBalances" error: ` + response.Message)
			t.FailNow()
		}

		response = stub.MockInvoke("1", [][]byte{[]byte("getBalances")})
		if response.Status >= 400 {
			fmt.Println(`"getBalances" error: ` + response.Message)
			t.FailNow()
		}

		var results []queryResult
		if err := json.Unmarshal(response.Payload, &results); err != nil {
			fmt.Println("JSON unmarshalling error: " + err.Error())
			t.FailNow()
		}

		if err := checkBalanceQuery(results, expectedResults); err != nil {
			fmt.Println(err)
			t.FailNow()
		}
	}

	// trying to remove balances without nsd permissions must lead to an error
	stub.SetCaller("org1")
	response = stub.MockInvoke("1", [][]byte{[]byte("removeBalances"), []byte(`[{
			"organization": "org1",
			"balances": [
				{
					"account": "MZ0987654321",
					"division": "19000000000000000"
				}
			]
		}]`)})
	if response.Status < 400 {
		fmt.Println(`"removeBalances" can be called without nsd permissions.`)
		t.FailNow()
	} else {
		fmt.Println(`Predicted "removeBalances" error: ` + response.Message)
	}
}

func TestCreateAlamedaFopXMLs(t *testing.T) {
	instruction := nsd.Instruction{
		Key: nsd.InstructionKey{
			Transferer: nsd.Balance{
				Account: "transf_acc",
				Division: "transf_div",
			},
			Receiver: nsd.Balance{
				Account: "recv_acc",
				Division: "recv_div",
			},
			Security: "RU000A0JVVB5",
			Quantity: "500",
			Reference: "SOMEREF123",
			InstructionDate: "2018-03-29",
			TradeDate: "2018-03-29",
			Type: nsd.InstructionTypeFOP,
		},

		Value: nsd.InstructionValue{
			DeponentFrom: "MCXXXXX00000",
			DeponentTo: "MSYYYYY00000",
			Status: "matched",
			Initiator: "tranferer",
			MemberInstructionIdFrom: "id_from",
			MemberInstructionIdTo: "id_to",
			ReasonFrom: nsd.Reason{
				Description: "Doc_from",
				Document: "123",
				DocumentDate: "2018-03-29",
			},
			ReasonTo: nsd.Reason{
				Description: "Doc_to",
				Document: "321",
				DocumentDate: "2018-03-29",
			},
		},
	}

	from, to := CreateAlamedaXMLsTestWrapper(&instruction, nsd.InstructionTypeFOP)

	const fromExpected = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>NDC000000000</deposit_c>
<contrag_c>MCXXXXX00000</contrag_c>
<contr_d_id>id_from</contr_d_id>
<createdate>2018-03-29</createdate>
<order_t_id>16</order_t_id>
<execute_dt>2018-03-29 00:00:00</execute_dt>
<expirat_dt>2018-04-27 23:59:59</expirat_dt>
</ORDER_HEADER>
<MF010>
<dep_acc_c>transf_acc</dep_acc_c>
<sec_c>transf_div</sec_c>
<deponent_c>MCXXXXX00000</deponent_c>
<corr_acc_c>recv_acc</corr_acc_c>
<corr_sec_c>recv_div</corr_sec_c>
<corr_code>MSYYYYY00000</corr_code>
<based_on>Doc_from</based_on>
<based_numb>123</based_numb>
<based_date>2018-03-29</based_date>
<securities>
<security>
<security_c>RU000A0JVVB5</security_c>
<security_q>500</security_q>
</security>
</securities>
<deal_reference>SOMEREF123</deal_reference>
<date_deal>2018-03-29</date_deal>
</MF010>
</Document>
</Batch>`

	const toExpected = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>NDC000000000</deposit_c>
<contrag_c>MSYYYYY00000</contrag_c>
<contr_d_id>id_to</contr_d_id>
<createdate>2018-03-29</createdate>
<order_t_id>16/1</order_t_id>
<execute_dt>2018-03-29 00:00:00</execute_dt>
<expirat_dt>2018-04-27 23:59:59</expirat_dt>
</ORDER_HEADER>
<MF010>
<dep_acc_c>transf_acc</dep_acc_c>
<sec_c>transf_div</sec_c>
<deponent_c>MCXXXXX00000</deponent_c>
<corr_acc_c>recv_acc</corr_acc_c>
<corr_sec_c>recv_div</corr_sec_c>
<corr_code>MSYYYYY00000</corr_code>
<based_on>Doc_to</based_on>
<based_numb>321</based_numb>
<based_date>2018-03-29</based_date>
<securities>
<security>
<security_c>RU000A0JVVB5</security_c>
<security_q>500</security_q>
</security>
</securities>
<deal_reference>SOMEREF123</deal_reference>
<date_deal>2018-03-29</date_deal>
</MF010>
</Document>
</Batch>`

	if from != fromExpected {
		t.Errorf("XML \"from\"is not equal expected value")
		fmt.Println(from)
	}
	if to != toExpected {
		t.Errorf("XML \"to\"is not equal expected value")
		fmt.Println(to)
	}
}

func TestCreateAlamedaDvpXMLs(t *testing.T) {
	instruction := nsd.Instruction{
		Key: nsd.InstructionKey{
			Transferer: nsd.Balance{
				Account: "transf_acc",
				Division: "transf_div",
			},
			Receiver: nsd.Balance{
				Account: "recv_acc",
				Division: "recv_div",
			},
			Security: "RU000A0JVVB5",
			Quantity: "500",
			Reference: "SOMEREF123",
			InstructionDate: "2018-03-29",
			TradeDate: "2018-03-29",
			Type: nsd.InstructionTypeDVP,
			TransfererRequisites: nsd.Requisites{
				Account: "tr_money_acc",
				Bic: "tr_money_bic",
			},
			ReceiverRequisites: nsd.Requisites{
				Account: "rc_money_acc",
				Bic: "rc_money_bic",
			},
			PaymentAmount: "10000.00",
			PaymentCurrency: "RUB",
		},

		Value: nsd.InstructionValue{
			DeponentFrom: "MCXXXXX00000",
			DeponentTo: "MSYYYYY00000",
			Status: "matched",
			Initiator: "tranferer",
			MemberInstructionIdFrom: "id_from",
			MemberInstructionIdTo: "id_to",
			ReasonFrom: nsd.Reason{
				Description: "Doc_from",
				Document: "123",
				DocumentDate: "2018-03-29",
			},
			ReasonTo: nsd.Reason{
				Description: "Doc_to",
				Document: "321",
				DocumentDate: "2018-03-29",
			},
			AdditionalInformation: nsd.Reason{
				Description: "additional info",
			},
		},
	}

	from, to := CreateAlamedaXMLsTestWrapper(&instruction, nsd.InstructionTypeDVP)

	const fromExpected = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>NDC000000000</deposit_c>
<contrag_c>MCXXXXX00000</contrag_c>
<contr_d_id>id_from</contr_d_id>
<createdate>2018-03-29</createdate>
<order_t_id>16/2</order_t_id>
<execute_dt>2018-03-29 00:00:00</execute_dt>
<expirat_dt>2018-03-29 23:59:59</expirat_dt>
</ORDER_HEADER>
<MF170>
<dep_acc_c>transf_acc</dep_acc_c>
<sec_c>transf_div</sec_c>
<corr_acc_c>recv_acc</corr_acc_c>
<corr_sec_c>recv_div</corr_sec_c>
<deal_num>SOMEREF123</deal_num>
<deal_date>2018-03-29</deal_date>
<con_code>MSYYYYY00000</con_code>
<sen_acc>rc_money_acc</sen_acc>
<sen_bic>rc_money_bic</sen_bic>
<rec_acc>tr_money_acc</rec_acc>
<rec_bic>tr_money_bic</rec_bic>
<pay_sum>10000.00</pay_sum>
<pay_curr>RUB</pay_curr>
<based_on>Doc_from</based_on>
<block_securities>N</block_securities>
<f_instruction>N</f_instruction>
<auto_borr>N</auto_borr>
<securities>
<security>
<security_c>RU000A0JVVB5</security_c>
<security_q>500</security_q>
</security>
</securities>
</MF170>
</Document>
</Batch>`

	const toExpected = `<?xml version="1.0" encoding="Windows-1251"?>
<Batch>
<Documents_amount>1</Documents_amount>
<Document DOC_ID="1" version="7">
<ORDER_HEADER>
<deposit_c>NDC000000000</deposit_c>
<contrag_c>MSYYYYY00000</contrag_c>
<contr_d_id>id_to</contr_d_id>
<createdate>2018-03-29</createdate>
<order_t_id>16/3</order_t_id>
<execute_dt>2018-03-29 00:00:00</execute_dt>
<expirat_dt>2018-03-29 23:59:59</expirat_dt>
</ORDER_HEADER>
<MF170>
<dep_acc_c>transf_acc</dep_acc_c>
<sec_c>transf_div</sec_c>
<corr_acc_c>recv_acc</corr_acc_c>
<corr_sec_c>recv_div</corr_sec_c>
<deal_num>SOMEREF123</deal_num>
<deal_date>2018-03-29</deal_date>
<con_code>MCXXXXX00000</con_code>
<sen_acc>rc_money_acc</sen_acc>
<sen_bic>rc_money_bic</sen_bic>
<rec_acc>tr_money_acc</rec_acc>
<rec_bic>tr_money_bic</rec_bic>
<pay_sum>10000.00</pay_sum>
<pay_curr>RUB</pay_curr>
<based_on>Doc_to</based_on>
<block_securities>N</block_securities>
<f_instruction>Y</f_instruction>
<auto_borr>N</auto_borr>
<add_info>/NZP additional info</add_info>
<securities>
<security>
<security_c>RU000A0JVVB5</security_c>
<security_q>500</security_q>
</security>
</securities>
</MF170>
</Document>
</Batch>`

	if from != fromExpected {
		t.Errorf("XML \"from\"is not equal expected value")
		fmt.Println(from)
	}
	if to != toExpected {
		t.Errorf("XML \"to\"is not equal expected value")
		fmt.Println(to)
	}
}