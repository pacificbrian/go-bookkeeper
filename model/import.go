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
	"time"
	"github.com/aclindsa/ofxgo"
	"github.com/pacificbrian/qif"
	"github.com/shopspring/decimal"
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
	c.Transnum = string(ofxTran.FiTID)
	c.PayeeName = string(ofxTran.Name)
	TrnAmt, _ := ofxTran.TrnAmt.Float64()
	c.Amount = decimal.NewFromFloatWithExponent(TrnAmt, -2)
	c.Memo = string(ofxTran.Memo)
}

func (c *CashFlow) makeCashFlowQIF(qifTran qif.BankingTransaction) {
	c.Date = qifTran.Date()
	c.setDefaults() // needs c.Date
	c.Transnum = qifTran.Num()
	c.PayeeName = qifTran.Payee()
	c.Amount = qifTran.AmountDecimal()
	c.Memo = qifTran.Memo()
}

func (t *Trade) makeTradeQIF(qifTran qif.InvestmentTransaction) {
	t.TradeTypeID = actionToTradeType(qifTran.Action())
	t.Date = qifTran.Date()
	t.Amount = qifTran.AmountDecimal()
	t.Shares = qifTran.Shares()
	t.Price = qifTran.Price()
	t.setDefaults() // needs t.Date, t.Shares
}

func (im *Import) create(db *gorm.DB) error {
	im.CreatedOn = time.Now()
	result := db.Omit(clause.Associations).Create(im)
	return result.Error
}

func (im *Import) ImportFile(session *Session, importFile HttpFile) error {
	fileName := importFile.FileName
	fileExtension := filepath.Ext(fileName)
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
	count = len(transactions)
	if count == 0 {
		return nil
	}
	spewModel(transactions)

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
			cashflows[i].insertCashFlow(db)
			entered++
		}
	case qif.TransactionTypeInvestment:
		if !im.Account.IsInvestment() {
			break
		}
		trades := make([]Trade, count)

		for i := 0; i < count; i++ {
			transaction := transactions[i].(qif.InvestmentTransaction)
			securityName := transaction.SecurityName()
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

	log.Printf("[MODEL] IMPORT(%d) [%s] QIF TRANSACTIONS (ACCEPTED %d of %d)",
		   im.ID, fileName, entered, count)
	return nil
}

func (im *Import) ImportFromQFX(session *Session, importFile HttpFile) error {
	var ofxTran []ofxgo.Transaction
	var entries []CashFlow
	fileName := importFile.FileName
	db := session.DebugDB
	count := 0

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

	if len(resp.Bank) > 0 {
		stmt, valid := resp.Bank[0].(*ofxgo.StatementResponse)
		if valid {
			ofxTran = stmt.BankTranList.Transactions
			count = len(ofxTran)
			entries = make([]CashFlow, count)

			// write Import, we store ImportID in CashFlows
			if count > 0 {
				im.create(db)
			}
		}
	}

	// convert ofxgo response to CashFlows
	for i := 0; i < count; i++ {
		entries[i].makeCashFlowOFX(&ofxTran[i])
		entries[i].AccountID = im.Account.ID
		entries[i].Account.cloneVerified(&im.Account)
		entries[i].ImportID = im.ID
		entries[i].insertCashFlow(db)
	}

	log.Printf("[MODEL] IMPORT(%d) [%s] OFX TRANSACTIONS (%d)",
		   im.ID, fileName, count)
	return nil
}
