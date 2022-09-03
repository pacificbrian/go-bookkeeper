/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type ModelWithDelete struct {
        ID        uint `gorm:"primaryKey"`
        DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Model struct {
        ID        uint `gorm:"primaryKey"`
        //DeletedAt gorm.DeletedAt `gorm:"index"`
}
