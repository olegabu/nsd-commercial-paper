package main

import (
	"testing"
	"github.com/olegabu/nsd-commercial-paper-common"
	"fmt"
)

func TestCreateAlamedaDvpXMLsTestWrapper(t *testing.T) {
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
			Type: "dvp",
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

	from, to := CreateAlamedaDvpXMLsTestWrapper(&instruction)

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
<based_numb>123</based_numb>
<based_date>2018-03-29</based_date>
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
</Batch>
`

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
<based_numb>321</based_numb>
<based_date>2018-03-29</based_date>
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
</Batch>
`

	if from != fromExpected {
		t.Errorf("XML \"from\"is not equal expected value")
		fmt.Println(from)
	}
	if to != toExpected {
		t.Errorf("XML \"to\"is not equal expected value")
		fmt.Println(to)
	}
}