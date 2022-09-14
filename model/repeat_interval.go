/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/shopspring/decimal"
)

type RepeatInterval struct {
	Model
	CashFlowID uint
	RepeatIntervalTypeID uint `form:"repeat_interval_type_id"`
	RepeatIntervalType RepeatIntervalType
	RepeatsLeft uint `form:"repeats"`
	Rate decimal.Decimal `form:"amount"`
}
