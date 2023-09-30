/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"encoding/json"
	"log"
	"github.com/shopspring/decimal"
	"github.com/pacificbrian/go-bookkeeper/config"
)

type SecurityChartDataSet struct {
	Label string `json:"label"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor string `json:"borderColor"`
	Fill string `json:"fill"`
	Data []decimal.Decimal `json:"data"`
}

type SecurityChart struct {
	Months []string `json:"labels"`
	Datasets []*SecurityChartDataSet `json:"datasets"`
}

type SecurityChartTickOptions struct {
}

type SecurityChartAxisOptions struct {
	Ticks SecurityChartTickOptions `json:"ticks"`
}

type SecurityChartScaleOptions struct {
	//YAxis SecurityChartAxisOptions `json:"yAxis"`
}

type SecurityChartOptions struct {
	Scales SecurityChartScaleOptions `json:"scales"`
}

func (s *Security) ChartsEnabled() bool {
	globals := config.GlobalConfig()
	return globals.EnableSecurityCharts
}

func (s *Security) GetChartOptionsByte() []byte {
	opts := new(SecurityChartOptions)
	//yAxis := &opts.Scales.YAxis

	jsonData, err := json.Marshal(opts)
	if err != nil {
		log.Println(err)
	}
	return jsonData
}

func (s *Security) GetChartOptions() string {
	return string(s.GetChartOptionsByte())
}

func (s *Security) GetChartDataByte(days int) []byte {
	dataset0 := new(SecurityChartDataSet)
	prices := []decimal.Decimal{decimal.NewFromInt32(10), decimal.NewFromInt32(25)}
	dataset0.Data = prices
	dataset0.Label = s.Company.GetName()
	dataset0.BorderColor = "#3B82F6"
	dataset0.BackgroundColor = "transparent"
	dataset0.Fill = "origin"

	data := new(SecurityChart)
	data.Months = append(data.Months, "January", "February")
	data.Datasets = append(data.Datasets, dataset0)

	log.Printf("[MODEL] SECURITY(%d) CHART DATA DAYS(%d)", s.ID, days)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	return jsonData
}

func (s *Security) GetChartData(days int) string {
	return string(s.GetChartDataByte(days))
}
