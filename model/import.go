/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"path/filepath"
	"fmt"
	"log"
	"strings"
	"time"
	"github.com/aclindsa/ofxgo"
	"github.com/pacificbrian/qif"
	"github.com/shopspring/decimal"
	"github.com/pacificbrian/go-bookkeeper/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Import struct {
	Model
	AccountID uint `gorm:"not null"`
	CashFlowCount uint `gorm:"-:all"`
	TradeCount uint `gorm:"-:all"`
	Username string `gorm:"-:all" form:"import.Username"`
	Password string `gorm:"-:all" form:"import.Password"`
	CreatedOn time.Time
	Account Account
}

var payeeNameLength int = 22

// i.Account must be preloaded
func (i *Import) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	i.Account.Verified = !(u == nil || i.Account.ID == 0 || u.ID != i.Account.UserID)
	if i.Account.Verified {
		i.Account.User = *u
		i.Account.Session = session
	}
	return i.Account.Verified
}

func (im *Import) Get(session *Session) *Import {
	db := session.DB
	// Verify we have access to Import
	if im.ID > 0 {
		db.Preload("Account").First(&im)
	}
	if !im.HaveAccessPermission(session) {
		return nil
	}
	return im
}

func (im *Import) ListImported(session *Session) []CashFlow {
	entries := []CashFlow{}
	if !im.Account.Verified {
		return entries
	}
	priorBalance := decimal.Zero
	db := session.DB

	if im.Account.IsInvestment() {
		entries = new(Trade).ListImportedCashFlows(im)
	}
	if (len(entries) == 0) {
		db.Order("date asc").Preload("Payee").
		   Where(&CashFlow{AccountID: im.AccountID, ImportID: im.ID}).
		   Find(&entries)
	}

	for i := 0; i < len(entries); i++ {
		c := &entries[i]
		c.Account.cloneVerified(&im.Account)
		c.Preload(db)
		c.Balance = priorBalance.Add(c.Amount)
		priorBalance = c.Balance
	}

	log.Printf("[MODEL] LIST IMPORT(%d) CASHFLOWS ACCOUNT(%d:%d)",
		   im.ID, im.AccountID, len(entries))
	return entries
}

func (im *Import) CountImported(session *Session) {
	var count int64 = 0
	if !im.Account.Verified {
		return
	}
	db := session.DB

	if im.Account.IsInvestment() {
		db.Model(&Trade{}).
		   Where(&Trade{AccountID: im.AccountID, ImportID: im.ID}).
		   Count(&count)
		im.TradeCount = uint(count)
		log.Printf("[MODEL] COUNT IMPORT(%d) TRADES ACCOUNT(%d:%d)",
			   im.ID, im.AccountID, im.TradeCount)

	}
	db.Model(&CashFlow{}).
	   Where(&CashFlow{AccountID: im.AccountID, ImportID: im.ID}).
	   Count(&count)
	im.CashFlowCount = uint(count)
	log.Printf("[MODEL] COUNT IMPORT(%d) CASHFLOWS ACCOUNT(%d:%d)",
		   im.ID, im.AccountID, im.CashFlowCount)
}

func dateFromOFX(ofxTran *ofxgo.Transaction) time.Time {
	//dateStr := ofxTran.DtPosted.String()
	return time.Date(ofxTran.DtPosted.Year(),
			 ofxTran.DtPosted.Month(),
			 ofxTran.DtPosted.Day(),
			 ofxTran.DtPosted.Hour(),
			 ofxTran.DtPosted.Minute(),
			 ofxTran.DtPosted.Second(),
			 ofxTran.DtPosted.Nanosecond(),
			 ofxTran.DtPosted.Location())
}

func (c *CashFlow) makeCashFlowOFX(ofxTran *ofxgo.Transaction) {
	globals := config.GlobalConfig()

	c.Date = dateFromOFX(ofxTran)
	c.setDefaults() // needs c.Date
	TrnAmt, _ := ofxTran.TrnAmt.Float64()
	c.Amount = decimal.NewFromFloatWithExponent(TrnAmt, -2)

	tranName := string(ofxTran.Name)
	if globals.LimitImportPayeeNameLength &&
	   len(tranName) > payeeNameLength+1 && tranName[payeeNameLength] == ' ' {
		c.Payee.Name = strings.TrimSpace(tranName[0:payeeNameLength+1])
		c.Payee.Address = strings.TrimSpace(tranName[payeeNameLength+1:])
	} else {
		c.Payee.Name = strings.TrimSpace(tranName)
	}
	c.PayeeName = c.Payee.Name
	c.Memo = strings.TrimSpace(string(ofxTran.Memo))
	c.Transnum = strings.TrimSpace(string(ofxTran.FiTID))
}

func (c *CashFlow) makeCashFlowQIF(qifTran qif.BankingTransaction) {
	c.Date = qifTran.Date()
	c.setDefaults() // needs c.Date
	c.Amount = qifTran.AmountDecimal()
	c.Payee.Name = strings.TrimSpace(qifTran.Payee())
	c.PayeeName = c.Payee.Name
	c.Memo = strings.TrimSpace(qifTran.Memo())
	c.Transnum = strings.TrimSpace(qifTran.Num())
}

func (t *Trade) applyTradeFixups(memo string) {
	globals := config.GlobalConfig()

	if !globals.EnableImportTradeFixups {
		return
	}

	// workarounds which have found to be needed
	switch memo {
	case "Change in Market Value":
		t.TradeTypeID = ReinvestedDividend
	case "Fees/Credits":
		t.TradeTypeID = ReinvestedDividend
	}
}

func (t *Trade) makeTradeQIF(qifTran qif.InvestmentTransaction) {
	memo := strings.TrimSpace(qifTran.Memo())
	t.TradeTypeID = actionToTradeType(qifTran.Action())
	t.Date = qifTran.Date()
	t.Amount = qifTran.AmountDecimal()
	t.Shares = qifTran.Shares()
	t.Price = qifTran.Price()
	t.applyTradeFixups(memo)
	t.setDefaults() // needs t.Date, t.Shares
}

func (im *Import) create(db *gorm.DB) error {
	im.CreatedOn = time.Now()
	result := db.Omit(clause.Associations).Create(im)
	return result.Error
}

func (im *Import) FetchOFX(session *Session) error {
	if (im.Username == "") {
		return errors.New("Invalid Username")
	}
	if (im.Password == "") {
		return errors.New("Invalid Password")
	}
	db := session.DB

	// Verify we have access to Account
	if !im.Account.Verified {
		im.Account.ID = im.AccountID
		account := im.Account.Get(session, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	if im.Account.InstitutionID < 1 {
		return errors.New("No OFX Institution")
	}

	im.Account.Institution.ID = im.Account.InstitutionID
	db.First(&im.Account.Institution)
	lastCF := im.Account.lastCashFlow(false)
	log.Printf("[MODEL] FETCH OFX FOR INSTITUTION(%s:%d) AFTER(%s)",
		   im.Account.Institution.Name, im.Account.OfxIndex,
		   timeToString(&lastCF.Date))

	client := im.Account.Institution.getClient()
	uid,_ := ofxgo.RandomUID()
	var query ofxgo.Request

	im.setSignon(&query)
	query.SetClientFields(client)

	// Signup/AcctInfoRequest to get account number
	acctInfoRequest := ofxgo.AcctInfoRequest{ TrnUID: *uid }
	query.Signup = append(query.Signup, &acctInfoRequest)
	resp, err := client.Request(&query)
	if err != nil {
		errStr := fmt.Sprintf("[MODEL] FETCH OFX: bad Signup/acctInfo response: %v",
				      err)
		return errors.New(errStr)
	}
	spewModel(resp)
	query.Signup = nil

	err = im.checkResponse(resp, false)
	if err != nil {
		errStr := fmt.Sprintf("[MODEL] FETCH OFX: Signup response error: %v", err)
		return errors.New(errStr)
	}

	err = im.setStatementRequest(&query, uid, &lastCF.Date, resp)
	if err != nil {
		errStr := fmt.Sprintf("[MODEL] FETCH OFX: acctInfo error: %v",
				      err)
		return errors.New(errStr)
	}
	spewModel(query)

	resp, err = client.Request(&query)
	if err != nil {
		errStr := fmt.Sprintf("[MODEL] FETCH OFX: bad stmtRequest response: %v",
				      err)
		return errors.New(errStr)
	}
	spewModel(resp)

	err = im.checkResponse(resp, true)
	if err != nil {
		errStr := fmt.Sprintf("[MODEL] FETCH OFX: stmtRequest response error: %v",
				      err)
		return errors.New(errStr)
	}
	return im.ImportFromOFX(session, resp, &lastCF.Date)
}

func (im *Import) ImportFile(session *Session, importFile HttpFile) error {
	fileName := importFile.FileName
	fileExtension := strings.ToLower(filepath.Ext(fileName))
	if fileExtension == ".qif" {
		return im.ImportFromQIF(session, importFile)
	} else if fileExtension == ".qfx" || fileExtension == ".ofx" {
		return im.ImportFromQFX(session, importFile)
	}
	return errors.New(fmt.Sprintf("[MODEL] IMPORT [%s]: unsupported file type",
				      fileName))
}

func isReversedQIF(transactions []qif.Transaction, transactionType qif.TransactionType) bool {
	count := len(transactions)
	reversed := false

	if count > 1 {
		var dateStart, dateEnd time.Time

		switch transactionType {
		case qif.TransactionTypeBanking:
			dateStart = transactions[0].(qif.BankingTransaction).Date()
			dateEnd = transactions[count - 1].(qif.BankingTransaction).Date()
			break
		case qif.TransactionTypeInvestment:
			dateStart = transactions[0].(qif.InvestmentTransaction).Date()
			dateEnd = transactions[count - 1].(qif.InvestmentTransaction).Date()
		}
		reversed = dateStart.After(dateEnd)
	}
	return reversed
}

func (im *Import) ImportFromQIF(session *Session, importFile HttpFile) error {
	var transactions []qif.Transaction
	var idx, idxIncrement int
	fileName := importFile.FileName
	db := session.DebugDB
	recordImport := true
	count := 0
	entered := 0

	// Verify we have access to Account
	if !im.Account.Verified {
		im.Account.ID = im.AccountID
		account := im.Account.Get(session, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	r := qif.NewReader(importFile.FileData)
	transactions, err := r.ReadAll()
	if err != nil {
		return errors.New(fmt.Sprintf("[MODEL] IMPORT [%s]: error: %v",
				              fileName, err))
	}
	spewModel(transactions)

	count = len(transactions)
	if count == 0 || !recordImport {
		goto done
	}

	if isReversedQIF(transactions, r.ReadTransactionType()) {
		idx = count - 1
		idxIncrement = -1
	} else {
		idx = 0
		idxIncrement = 1
	}

	// convert qif.Transactions to CashFlows or Trades
	switch r.ReadTransactionType() {
	case qif.TransactionTypeBanking:
		cashflows := make([]CashFlow, count)

		// write Import, we store ImportID in CashFlows
		im.create(db)

		for i := 0; i < count; i++ {
			transaction := transactions[idx].(qif.BankingTransaction)
			idx = idx + idxIncrement
			cashflows[i].makeCashFlowQIF(transaction)
			cashflows[i].AccountID = im.Account.ID
			cashflows[i].Account.cloneVerified(&im.Account)
			cashflows[i].ImportID = im.ID
			if cashflows[i].insertCashFlow(db, true) == nil {
				entered++
			}
		}
	case qif.TransactionTypeInvestment:
		if !im.Account.IsInvestment() {
			break
		}
		trades := make([]Trade, count)

		for i := 0; i < count; i++ {
			transaction := transactions[idx].(qif.InvestmentTransaction)
			idx = idx + idxIncrement
			securityName := strings.TrimSpace(transaction.SecurityName())
			security := im.Account.securityGetByImportName(session,
								       securityName)
			if security == nil {
				continue
			} else if im.ID == 0 {
				// write Import, we store ImportID in CashFlows
				im.create(db)
			}

			trades[i].makeTradeQIF(transaction)
			trades[i].SecurityID = security.ID
			trades[i].ImportID = im.ID
			trades[i].insertTrade(db, security)
			entered++
		}
	default:
		count = 0
	}

done:
	log.Printf("[MODEL] IMPORT(%d) [%s] QIF TRANSACTIONS (ACCEPTED %d of %d)",
		   im.ID, fileName, entered, count)
	return nil
}

func isReversedOFX(ofxTran []ofxgo.Transaction) bool {
	count := len(ofxTran)
	reversed := false

	if count > 1 {
		dateStart := dateFromOFX(&ofxTran[0])
		dateEnd := dateFromOFX(&ofxTran[count - 1])
		reversed = dateStart.After(dateEnd)
	}
	return reversed
}

func (im *Import) ImportFromQFX(session *Session, importFile HttpFile) error {
	fileName := importFile.FileName

	// Verify we have access to Account
	if !im.Account.Verified {
		im.Account.ID = im.AccountID
		account := im.Account.Get(session, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	resp, err := ofxgo.ParseResponse(importFile.FileData)
	if err != nil {
		return errors.New(fmt.Sprintf("[MODEL] IMPORT [%s]: bad response: %v",
				              fileName, err))
	}
	return im.ImportFromOFX(session, resp, nil)
}

func (im *Import) ImportFromOFX(session *Session, resp *ofxgo.Response, after *time.Time) error {
	var ofxTran []ofxgo.Transaction
	var entries []CashFlow
	db := session.DebugDB
	recordImport := true
	count := 0
	entered := 0

	ofxTran = im.getOfxTransactions(resp)
	count = len(ofxTran)
	spewModel(ofxTran)

	// write Import, we store ImportID in CashFlows
	if recordImport && count > 0 {
		var idx, idxIncrement int

		entries = make([]CashFlow, count)
		im.create(db)

		if isReversedOFX(ofxTran) {
			idx = count - 1
			idxIncrement = -1
		} else {
			idx = 0
			idxIncrement = 1
		}

		// convert ofxgo response to CashFlows
		for i := 0; i < count; i++ {
			entries[i].makeCashFlowOFX(&ofxTran[idx])
			idx = idx + idxIncrement
			if after != nil && !entries[i].Date.After(*after) {
				continue
			}
			entries[i].AccountID = im.Account.ID
			entries[i].Account.cloneVerified(&im.Account)
			entries[i].ImportID = im.ID
			if entries[i].insertCashFlow(db, true) == nil {
				entered++
			}
		}
	}

	log.Printf("[MODEL] IMPORT(%d) OFX TRANSACTIONS (ACCEPTED %d of %d)",
		   im.ID, entered, count)
	return nil
}
