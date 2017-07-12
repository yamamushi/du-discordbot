package main

// LoanCalculator struct
type LoanCalculator struct {
	bank *Bank
}

// LoanRecord struct
type LoanRecord struct {
	ID        string `storm:"id"`
	AccountID string `storm:"index"`
}

// GetLoanRecords function
func (h *LoanCalculator) GetLoanRecords(accountid string) {

}
