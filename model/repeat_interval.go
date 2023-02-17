/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"strconv"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepeatIntervalType struct {
	Model
	Name string `form:"repeat_interval_type.Name"`
	Days uint
}

type RepeatInterval struct {
	Model
	CashFlowID uint
	Rate decimal.Decimal
	RepeatIntervalTypeID uint `form:"repeat_interval_type_id"`
	RepeatsLeft uint `form:"repeats" gorm:"default:NULL"`
	RepeatsLeftPtr *uint `gorm:"-:all"`
	StartDay int
	RepeatIntervalType RepeatIntervalType
}

func (r RepeatInterval) GetRate() string {
	if r.Rate.IsZero() {
		return ""
	}
	return r.Rate.StringFixed(3)
}

func (r RepeatInterval) GetRepeatsLeft() string {
	if r.RepeatsLeftPtr == nil {
		return ""
	}
	return strconv.Itoa(int(r.RepeatsLeft))
}

func (r RepeatInterval) HasRepeatsLeft() bool {
	if r.RepeatsLeftPtr == nil {
		return true
	}
	return (r.RepeatsLeft > 0)
}

func (r *RepeatInterval) SetRepeatsLeft(repeats string) {
	repeatsLeft, _ := strconv.Atoi(repeats)
	r.RepeatsLeft = uint(repeatsLeft)

	// for special handling to allow RepeatsLeft = NULL (unset)
	// in Update()
	if repeats != "" { // set
		r.RepeatsLeftPtr = &r.RepeatsLeft
	} else { // unset
		r.RepeatsLeftPtr = nil
	}
}

func (r *RepeatInterval) Preload(db *gorm.DB) {
	if r.CashFlowID == 0 {
		db.Preload("RepeatIntervalType").First(&r)
	}

	// special handling for RepeatsLeft == NULL (not set)
	nullTest := new(RepeatInterval)
	db.Where("repeats_left IS NOT NULL").First(&nullTest, r.ID)
	if nullTest.ID == r.ID { // NOT NULL (set)
		r.RepeatsLeftPtr = &r.RepeatsLeft
	}

	if r.RepeatIntervalType.Name == "" {
		// need userCache lookup
		r.RepeatIntervalType.ID = r.RepeatIntervalTypeID
		db.First(&r.RepeatIntervalType)
	}
}

// r should already been Preloaded
func (r *RepeatInterval) advance(db *gorm.DB) int {
	days := int(r.RepeatIntervalType.Days)

	// decrement RepeatsLeft
	if r.RepeatsLeft > 0 {
		r.RepeatsLeft -= 1
		updates := map[string]interface{}{"repeats_left": gorm.Expr("repeats_left - ?", 1)}
		db.Omit(clause.Associations).Model(r).
		   Select("repeats_left").Updates(updates)
	} else if days == 0 {
		// if IntervalType == Once, don't let it repeat
		db.Omit(clause.Associations).Model(r).
		   Select("repeats_left").Updates(RepeatInterval{RepeatsLeft: 0})
	}

	// use helper, can't test r.RepeatsLeft because unset/NULL == 0
	if !r.HasRepeatsLeft() {
		days = 0 // hit when looping until final Repeat
	}

	log.Printf("[MODEL] ADVANCE REPEAT_INTERVAL(%d) DAYS(%d) LEFT(%d)",
		   r.ID, days, r.RepeatsLeft)
	return days
}

func (*RepeatIntervalType) List(db *gorm.DB) []RepeatIntervalType {
	entries := []RepeatIntervalType{}
	db.Find(&entries)

	return entries
}

func (r *RepeatInterval) Create(db *gorm.DB, c *CashFlow) error {
	r.CashFlowID = c.ID
	r.StartDay = c.Date.Day()
	result := db.Omit(clause.Associations).Create(r)
	log.Printf("[MODEL] CREATE REPEAT_INTERVAL(%d) FOR CASHFLOW(%d)",
		   r.ID, c.ID)
	return result.Error
}

func (r *RepeatInterval) Update() error {
	db := getDbManager()

	result := db.Omit(clause.Associations).Save(r)
	if result.Error == nil && r.GetRepeatsLeft() == "" {
		updates := map[string]interface{}{"repeats_left": nil}
		result = db.Omit(clause.Associations).Model(r).
			     Select("repeats_left").Updates(updates)
	}
	log.Printf("[MODEL] UPDATE REPEAT_INTERVAL(%d) FOR CASHFLOW(%d)",
		   r.ID, r.CashFlowID)
	return result.Error
}
