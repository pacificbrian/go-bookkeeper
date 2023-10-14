/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"github.com/aclindsa/ofxgo"
)

type Institution struct {
	Model
	AppVer uint
	FiId uint
	AppId string
	FiOrg string
	FiUrl string
	Name string
}

func (*Institution) List() []Institution {
	db := getDbManager()
	zero_entry := Institution{}
	sub_entries := []Institution{}
	var entries []Institution

	entries = append(entries, zero_entry)
	db.Find(&sub_entries)
	entries = append(entries, sub_entries...)

	return entries
}

func (inst *Institution) getClient() *ofxgo.BasicClient {
	if (inst.AppId != "" && inst.AppVer > 0) {
		return &ofxgo.BasicClient{ AppID: inst.AppId,
					   AppVer: strconv.Itoa(int(inst.AppVer)) }
	}
	return &ofxgo.BasicClient{ AppID: "QWIN", AppVer: "2900" }
}

func (im *Import) getOfxTransactions(resp *ofxgo.Response) []ofxgo.Transaction {
	if len(resp.Bank) > 0 {
		stmt, valid := resp.Bank[0].(*ofxgo.StatementResponse)
		if valid {
			return stmt.BankTranList.Transactions
		}
	} else if len(resp.CreditCard) > 0 {
		stmt, valid := resp.CreditCard[0].(*ofxgo.CCStatementResponse)
		if valid {
			return stmt.BankTranList.Transactions
		}
	}
	return nil
}
func (im *Import) setSignon(query *ofxgo.Request) {
	inst := &im.Account.Institution
	query.URL = inst.FiUrl
	query.Signon.Org = ofxgo.String(inst.FiOrg)
	query.Signon.Fid = ofxgo.String(strconv.Itoa(int(inst.FiId)))

	query.Signon.ClientUID = ofxgo.UID(im.Account.ClientUID)
	query.Signon.UserID = ofxgo.String(im.Username)
	query.Signon.UserPass = ofxgo.String(im.Password)
}

func (im *Import) setStatementRequest(query *ofxgo.Request, uid *ofxgo.UID, date *time.Time, resp *ofxgo.Response) error {
	acctIdx := int(im.Account.OfxIndex)
	if len(resp.Signup[0].(*ofxgo.AcctInfoResponse).AcctInfo) < acctIdx+1 {
		return errors.New("no acctinfo received")
	}
	acctInfo := resp.Signup[0].(*ofxgo.AcctInfoResponse).AcctInfo[acctIdx]

	DtEnd := &query.Signon.DtClient
	days := daysBetweenDates(date, &DtEnd.Time, false)
	DtStart := &ofxgo.Date{Time: DtEnd.AddDate(0, 0, int(-days))}

	switch im.Account.AccountTypeID {
	case AccountTypeDeposit:
		if acctInfo.BankAcctInfo == nil {
			return errors.New("no bank acctinfo received")
		}
		statementRequest := ofxgo.StatementRequest {
			TrnUID: *uid,
			Include: true,
		}
		statementRequest.BankAcctFrom = acctInfo.BankAcctInfo.BankAcctFrom
		statementRequest.DtEnd = DtEnd
		statementRequest.DtStart = DtStart
		query.Bank = append(query.Bank, &statementRequest)
	case AccountTypeCreditCard:
		if acctInfo.CCAcctInfo == nil {
			return errors.New("no cc acctinfo received")
		}
		statementRequest := ofxgo.CCStatementRequest {
			TrnUID: *uid,
			Include: true,
		}
		statementRequest.CCAcctFrom = acctInfo.CCAcctInfo.CCAcctFrom
		statementRequest.DtEnd = DtEnd
		statementRequest.DtStart = DtStart
		query.CreditCard = append(query.CreditCard, &statementRequest)
	case AccountTypeInvestment:
		if acctInfo.InvAcctInfo == nil {
			return errors.New("no investment acctinfo received")
		}
		statementRequest := ofxgo.InvStatementRequest {
			TrnUID: *uid,
			Include:        true,
			//IncludeOO:      true,
			//IncludePos:     true,
			//IncludeBalance: true,
			//Include401K:    true,
			//Include401KBal: true,
		}
		statementRequest.InvAcctFrom = acctInfo.InvAcctInfo.InvAcctFrom
		statementRequest.DtEnd = DtEnd
		statementRequest.DtStart = DtStart
		query.InvStmt = append(query.InvStmt, &statementRequest)
	}
	return nil
}

func (im *Import) checkResponse(resp *ofxgo.Response, hasStmtRequest bool) error {
	var accountTypeID uint

	if resp.Signon.Status.Code != 0 {
		meaning, _ := resp.Signon.Status.CodeMeaning()
		errStr := fmt.Sprintf("nonzero signon status (%d: %s) with message: %s\n",
				      resp.Signon.Status.Code, meaning,
				      resp.Signon.Status.Message)
		return errors.New(errStr)
	}

	if hasStmtRequest {
		accountTypeID = im.Account.AccountTypeID
	}

	switch accountTypeID {
	case 0:
		if len(resp.Signup) < 1 {
			return errors.New("no signup message received")
		}
	case AccountTypeDeposit:
		if len(resp.Bank) < 1 {
			return errors.New("no banking messages received")
		}
	case AccountTypeCreditCard:
		if len(resp.CreditCard) < 1 {
			return errors.New("no credit card messages received")
		}
	case AccountTypeInvestment:
		if len(resp.InvStmt) < 1 {
			return errors.New("no investment messages received")
		}
	}
	return nil
}
