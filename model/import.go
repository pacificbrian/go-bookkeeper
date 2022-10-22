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
	"gorm.io/gorm"
)

type Import struct {
	Model
	AccountID uint `gorm:"not null"`
	Account Account
	CreatedOn time.Time
	CashFlowCount uint `gorm:"-:all"`
}

func (imp *Import) ImportFile(db *gorm.DB, importFile HttpFile) error {
	fileName := importFile.FileName
	count := 0

	// Verify we have access to Account
	if !imp.Account.Verified {
		imp.Account.ID = imp.AccountID
		account := imp.Account.Get(db, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	resp, err := ofxgo.ParseResponse(importFile.FileData)
	if err != nil {
		return errors.New(fmt.Sprintf("[MODEL] IMPORT [%s]: bad response: %v", fileName, err))
	}
	defer importFile.FileData.Close()

	// dump response for now
	ForceSpewModel(resp)
	log.Printf("[MODEL] IMPORT [%s] OFX TRANSACTIONS (%d)", fileName, count)

	return nil
}
