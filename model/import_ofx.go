/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

type Institution struct {
	Model
	AppVer uint
	FiId uint
	AppId string
	FiOrg string
	FiUrl string
	Name string
}

func (*Institution) List() []Institution {
	db := getDbManager()
	zero_entry := Institution{}
	sub_entries := []Institution{}
	var entries []Institution

	entries = append(entries, zero_entry)
	db.Find(&sub_entries)
	entries = append(entries, sub_entries...)

	return entries
}
