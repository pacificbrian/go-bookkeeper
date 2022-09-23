/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

type Company struct {
	Model
	Name string `form:"Name"`
	Symbol string `form:"Symbol"`
}
