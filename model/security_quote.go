/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"strconv"
	"time"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/quote"
	"github.com/shopspring/decimal"
)

var Months = [14]string{"",
			"Jan ",
			"Feb ",
			"Mar ",
			"Apr ",
			"May ",
			"Jun ",
			"Jul ",
			"Aug ",
			"Sep ",
			"Oct ",
			"Nov ",
			"Dec "}

type SecurityQuote struct {
	lastQuoted time.Time
	Price decimal.Decimal
}

type SecurityQuoteCache struct {
	Quotes map[string]SecurityQuote
}

var quotes *SecurityQuoteCache

func (sqc *SecurityQuoteCache) init() {
	sqc.Quotes = make(map[string]SecurityQuote)
}

func init() {
	quotes = new(SecurityQuoteCache)
	quotes.init()
}

func GetQuoteCache() *SecurityQuoteCache {
	return quotes
}

// Decide fetch policy, probably once per hour, if market is open.
// This can be skipped and quotes can alway be forced.
func fetchIsAllowed(last *time.Time, now *time.Time) bool {
	if last.IsZero() {
		return true
	}
	// for now, limit quotes to once per (business) day
	return (daysBetweenDates(last, now, true) >= 1)
}

func (sqc *SecurityQuoteCache) add(symbol string, quote *SecurityQuote) {
	log.Printf("[CACHE] ADD QUOTE FOR SYMBOL(%s)", symbol)
	sqc.Quotes[symbol] = *quote
}

func (sqc *SecurityQuoteCache) GetDateOf(symbol string) time.Time {
	return sqc.Quotes[symbol].lastQuoted
}

func (sqc *SecurityQuoteCache) Get(symbol string) SecurityQuote {
	return sqc.Quotes[symbol]
}

func (s *Security) fetchPrice(force bool) *SecurityQuote {
	var last *time.Time = &s.lastQuoteUpdate
	securityQuote := new(SecurityQuote)
	curTime := time.Now()

	if s.Company.Symbol == "" ||
	   GetQuoteCache() == nil ||
	   !SecurityTypeIsPriceFetchable[s.SecurityTypeID] {
		return nil
	}

	// determine time of last Quote in cache or database
	lastQuoted := GetQuoteCache().GetDateOf(s.Company.Symbol)
	if lastQuoted.After(*last) {
		last = &lastQuoted
	}

	if !force && !fetchIsAllowed(last, &curTime) {
		return nil
	}

	q, err := quote.Get(s.Company.Symbol)
	if err != nil {
		log.Println(err)
		return nil
	}
	if q == nil {
		return nil
	}
	spewModel(q)

	//price := q.RegularMarketPreviousClose
	price := q.RegularMarketPrice
	securityQuote.Price = decimal.NewFromFloatWithExponent(price, -3)
	securityQuote.lastQuoted = curTime
	log.Printf("[MODEL] SECURITY(%d) QUOTE SYMBOL(%s) PRICE(%f)",
		   s.ID, s.Company.Symbol, price)

	GetQuoteCache().add(s.Company.Symbol, securityQuote)
	return securityQuote
}

func (s *Security) fetchPrices(days int) ([]string, []decimal.Decimal) {
	labels := []string{}
	prices := []decimal.Decimal{}

	if s.Company.Symbol == "" {
		return labels, prices
	}
	p := &chart.Params{}
	p.Symbol = s.Company.Symbol
	p.Interval = datetime.OneDay

	t2 := time.Now()
	p.End = datetime.New(&t2)
	t1 := t2.AddDate(0,0,-days)
	p.Start = datetime.New(&t1)

	iter := chart.Get(p)
	for iter.Next() {
		b := iter.Bar()
		d := datetime.FromUnix(b.Timestamp)
		labels = append(labels, Months[d.Month] + strconv.Itoa(d.Day))
		prices = append(prices, b.AdjClose)
	}
	return labels, prices
}
