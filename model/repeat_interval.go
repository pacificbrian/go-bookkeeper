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

type RepeatInterval struct {
	Model
	CashFlowID uint
	RepeatIntervalTypeID uint `form:"repeat_interval_type_id"`
	RepeatIntervalType RepeatIntervalType
	RepeatsLeft uint `form:"repeats" gorm:"default:NULL"`
	RepeatsLeftPtr *uint `gorm:"-:all"`
	Rate decimal.Decimal `form:"rate"`
	StartDay int
}

func (r RepeatInterval) GetRepeatsLeft() string {
	if r.RepeatsLeftPtr == nil {
		return ""
	}
	return strconv.Itoa(int(r.RepeatsLeft))
}

func (r *RepeatInterval) Preload(db *gorm.DB) {
	//db.Preload("RepeatIntervalType").First(&r)
	db.First(&r)

	// special handling for RepeatsLeft == NULL (not set)
	nullTest := new(RepeatInterval)
	db.Where("repeats_left IS NOT NULL").First(&nullTest, r.ID)
	if nullTest.ID == r.ID {
		r.RepeatsLeftPtr = &r.RepeatsLeft
	}

	// need userCache lookup
	r.RepeatIntervalType.ID = r.RepeatIntervalTypeID
	db.First(&r.RepeatIntervalType)
}

// r should already been Preloaded
func (r *RepeatInterval) Advance(db *gorm.DB) int {
	days := int(r.RepeatIntervalType.Days)

	// decrement RepeatsLeft
	if r.RepeatsLeft > 0 {
		r.RepeatsLeft -= 1
		updates := map[string]interface{}{"repeats_left": gorm.Expr("repeats_left - ?", 1)}
		db.Omit(clause.Associations).Model(r).
		   Select("repeats_left").Updates(updates)
	}
	if r.RepeatsLeft == 0 {
		days = 0 // hit when looping until final Repeat
	}

	return days
}

func (r *RepeatInterval) Create(db *gorm.DB, c *CashFlow) error {
	r.CashFlowID = c.ID
	r.StartDay = c.Date.Day()
	result := db.Create(r)
	log.Printf("[MODEL] CREATE REPEAT_INTERVAL(%d) FOR CASHFLOW(%d)",
		   c.RepeatIntervalID, c.ID)
	return result.Error
}

func (r *RepeatInterval) Update(db *gorm.DB) error {
	result := db.Save(r)
	log.Printf("[MODEL] UPDATE REPEAT_INTERVAL(%d) FOR CASHFLOW(%d)",
		   r.ID, r.CashFlowID)
	return result.Error
}
