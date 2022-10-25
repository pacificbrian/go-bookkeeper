/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"time"
	"github.com/piquette/finance-go/quote"
	"github.com/shopspring/decimal"
)

type SecurityQuote struct {
	lastQuoted time.Time
	Price decimal.Decimal
}

// Decide fetch policy, probably once per hour, if market is open.
// This can be skipped and quotes can alway be forced.
func fetchIsAllowed(last *time.Time, now *time.Time) bool {
	// for now, limit quotes to once per (business) day
	return (daysBetweenDates(last, now, true) >= 1)
}

func (s *Security) fetchPrice(last *time.Time, force bool) *SecurityQuote {
	securityQuote := new(SecurityQuote)
	curTime := time.Now()

	if s.Company.Symbol == "" ||
	   !SecurityTypeIsPriceFetchable[s.SecurityTypeID] {
		return nil
	}

	if !force && !last.IsZero() && !fetchIsAllowed(last, &curTime) {
		return nil
	}

	q, err := quote.Get(s.Company.Symbol)
	if err != nil {
		log.Println(err)
		return nil
	}
	spewModel(q)

	securityQuote.Price = decimal.NewFromFloatWithExponent(q.RegularMarketPreviousClose, -3)
	securityQuote.lastQuoted = curTime
	log.Printf("[MODEL] SECURITY(%d) QUOTE SYMBOL(%s) PRICE(%f)",
		   s.ID, s.Company.Symbol, q.RegularMarketPreviousClose)

	return securityQuote
}
