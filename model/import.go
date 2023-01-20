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
	CreatedOn time.Time
	Account Account
}

func (c *CashFlow) makeCashFlowOFX(ofxTran *ofxgo.Transaction) {
	//dateStr := ofxTran.DtPosted.String()
	c.Date = time.Date(ofxTran.DtPosted.Year(),
			   ofxTran.DtPosted.Month(),
			   ofxTran.DtPosted.Day(),
			   ofxTran.DtPosted.Hour(),
			   ofxTran.DtPosted.Minute(),
			   ofxTran.DtPosted.Second(),
			   ofxTran.DtPosted.Nanosecond(),
			   ofxTran.DtPosted.Location())
	c.setDefaults() // needs c.Date
	c.Transnum = strings.TrimSpace(string(ofxTran.FiTID))
	c.PayeeName = strings.TrimSpace(string(ofxTran.Name))
	TrnAmt, _ := ofxTran.TrnAmt.Float64()
	c.Amount = decimal.NewFromFloatWithExponent(TrnAmt, -2)
	c.Memo = strings.TrimSpace(string(ofxTran.Memo))
}

func (c *CashFlow) makeCashFlowQIF(qifTran qif.BankingTransaction) {
	c.Date = qifTran.Date()
	c.setDefaults() // needs c.Date
	c.Transnum = strings.TrimSpace(qifTran.Num())
	c.PayeeName = strings.TrimSpace(qifTran.Payee())
	c.Amount = qifTran.AmountDecimal()
	c.Memo = strings.TrimSpace(qifTran.Memo())
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

func (im *Import) ImportFromQIF(session *Session, importFile HttpFile) error {
	var transactions []qif.Transaction
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

	// convert qif.Transactions to CashFlows or Trades
	switch transactions[0].TransactionType() {
	case qif.TransactionTypeBanking:
		cashflows := make([]CashFlow, count)

		// write Import, we store ImportID in CashFlows
		im.create(db)

		for i := 0; i < count; i++ {
			transaction := transactions[i].(qif.BankingTransaction)
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
			transaction := transactions[i].(qif.InvestmentTransaction)
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

func (im *Import) ImportFromQFX(session *Session, importFile HttpFile) error {
	var ofxTran []ofxgo.Transaction
	var entries []CashFlow
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

	resp, err := ofxgo.ParseResponse(importFile.FileData)
	if err != nil {
		return errors.New(fmt.Sprintf("[MODEL] IMPORT [%s]: bad response: %v",
				              fileName, err))
	}
	spewModel(ofxTran)

	if len(resp.Bank) > 0 {
		stmt, valid := resp.Bank[0].(*ofxgo.StatementResponse)
		if valid {
			ofxTran = stmt.BankTranList.Transactions
			count = len(ofxTran)
			entries = make([]CashFlow, count)
		}
	} else if len(resp.CreditCard) > 0 {
		stmt, valid := resp.CreditCard[0].(*ofxgo.CCStatementResponse)
		if valid {
			ofxTran = stmt.BankTranList.Transactions
			count = len(ofxTran)
			entries = make([]CashFlow, count)
		}
	}

	// write Import, we store ImportID in CashFlows
	if recordImport && count > 0 {
		im.create(db)

		// convert ofxgo response to CashFlows
		for i := 0; i < count; i++ {
			entries[i].makeCashFlowOFX(&ofxTran[i])
			entries[i].AccountID = im.Account.ID
			entries[i].Account.cloneVerified(&im.Account)
			entries[i].ImportID = im.ID
			if entries[i].insertCashFlow(db, true) == nil {
				entered++
			}
		}
	}

	log.Printf("[MODEL] IMPORT(%d) [%s] OFX TRANSACTIONS (ACCEPTED %d of %d)",
		   im.ID, fileName, entered, count)
	return nil
}
