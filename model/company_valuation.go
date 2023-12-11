/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	ROA uint = iota
	ROE
	ROIC
)

// order of below must match above for map keys to be correct
var ValuationMetricKeys = []string {
	"ROA",
	"ROE",
	"ROIC",
}

var ValuationMetricMap = map[string]func(*Company, FilingData) (float64, error){
	ValuationMetricKeys[ROA] : getReturnOnAssets,
	ValuationMetricKeys[ROE] : getReturnOnEquity,
	ValuationMetricKeys[ROIC] : getReturnOnInvestedCapital,
}

func (c *Company) GetValuationMetricNames(viewName string) []string {
	return ValuationMetricKeys
}

func getReturnOnAssets(c *Company, f FilingData) (float64, error){
	return 0.0, nil
}

func getReturnOnEquity(c *Company, f FilingData) (float64, error){
	return 0.0, nil
}

func getReturnOnInvestedCapital(c *Company, f FilingData) (float64, error){
	return 0.0, nil
}

func (c *Company) GetValuationMetric(f FilingData, item string) float64 {
	queryFunc := ValuationMetricMap[item]
	data, _ := queryFunc(c, f)
	return data
}

func (c *Company) GetValuationMetricString(f FilingData, item string) string {
	p := message.NewPrinter(language.English)
	data := c.GetValuationMetric(f, item)
	return p.Sprintf("%.2f", data) // add commas
}
