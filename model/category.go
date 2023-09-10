/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"gorm.io/gorm"
)

type CategoryType struct {
	Model
	Name string `form:"category_type.Name"`
}

type Category struct {
	Model
	UserID uint
	CategoryTypeID uint `form:"category.category_type_id"`
	Name string `form:"category.Name"`
	OmitFromPie bool
	CategoryType CategoryType
	User User
}

func (c *Category) IsInterest() bool {
	return (c.ID == 34)
}

func (c *Category) IsMortgageInterest() bool {
	return (c.ID == 35)
}

func (c *Category) IsInterestIncome() bool {
	return (c.ID == 74)
}

func (c *Category) LoanPI() bool {
	return (c.IsInterest() || c.IsMortgageInterest())
}

func CategoryGetByName(name string) *Category {
	db := getDbManager()

	c := new(Category)
	c.Name = name
	// need Where because name is not primary keys
	db.Where(&c).First(&c)

	log.Printf("[MODEL] GET CATEGORY(%d) BY NAME(%s)",
		   c.ID, name)
	return c
}

func (*CategoryType) List(db *gorm.DB) []CategoryType {
	entries := []CategoryType{}
	db.Find(&entries)

	return entries
}

func (*Category) List(db *gorm.DB) []Category {
	// need userCache lookup
	var entries []Category
	sub_entries := []Category{}

	// Expenses
	db.Order("Name").Where("(category_type_id < 2 OR category_type_id = 3)").
			 Find(&sub_entries)
	entries = append(entries, sub_entries...)

	// Income
	db.Order("Name").Where("(category_type_id = 0 OR category_type_id = 2)").
			 Find(&sub_entries)
	entries = append(entries, sub_entries...)

	log.Printf("[MODEL] LIST CATEGORIES (%d)", len(entries))
	return entries
}
