package main


type LoanCalculator struct {

	bank *Bank

}


type LoanRecord struct {

	ID	string `storm:"id"`
	AccountID	string `storm:"index"`

}

func (h *LoanCalculator) GetLoanRecords(accountid string){


}