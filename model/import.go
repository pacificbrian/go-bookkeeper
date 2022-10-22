/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/aclindsa/ofxgo"
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

func (c *CashFlow) makeCashFlow(ofxTran *ofxgo.Transaction) {
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

func (im *Import) create(db *gorm.DB) error {
	im.CreatedOn = time.Now()
	result := db.Omit(clause.Associations).Create(im)
	return result.Error
}

func (im *Import) ImportFile(db *gorm.DB, importFile HttpFile) error {
	var ofxTran []ofxgo.Transaction
	var entries []CashFlow
	fileName := importFile.FileName
	count := 0

	// Verify we have access to Account
	if !im.Account.Verified {
		im.Account.ID = im.AccountID
		account := im.Account.Get(db, false)
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
			im.create(db)
		}
	}

	// convert ofxgo response to CashFlows
	for i := 0; i < count; i++ {
		entries[i].makeCashFlow(&ofxTran[i])
		entries[i].AccountID = im.Account.ID
		entries[i].Account.cloneVerified(&im.Account)
		entries[i].ImportID = im.ID
		entries[i].insertCashFlow(db)
	}

	log.Printf("[MODEL] IMPORT(%d) [%s] OFX TRANSACTIONS (%d)", im.ID, fileName, count)
	return nil
}
