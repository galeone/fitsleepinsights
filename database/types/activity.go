// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"github.com/galeone/fitbit/v2/types"
)

type Category struct {
	types.Category
	ID            int64                 `igor:"primary_key"`
	Activities    []ActivityDescription `sql:"-"`
	SubCategories []SubCategory         `sql:"-"`
}

func (Category) TableName() string {
	return "categories"
}

type SubCategory struct {
	types.SubCategory
	ID int64 `igor:"primary_key"`
	// Overwrite the Category and SubCategory fields
	// with the foreign keys (unfortunately the name is without the _id suffix
	Category   int64
	Activities []ActivityDescription `sql:"-"`
}

func (SubCategory) TableName() string {
	return "subcategories"
}

type ActivityDescription struct {
	types.ActivityDescription
	ID int64 `igor:"primary_key"`

	// Overwrite the Category and SubCategory fields
	// with the foreign keys (unfortunately the name is without the _id suffix)
	Subcategory    int64
	Category       int64
	ActivityLevels []ActivityLevel `sql:"-"`
}

func (ActivityDescription) TableName() string {
	return "activities_descriptions"
}

type ActivityLevel struct {
	types.ActivityLevel
	ID int64 `igor:"primary_key"`
	// Overwrite the ActivityDescription field
	// with the foreign keys (unfortunately the name is without the _id suffix
	ActivityDescription int64
}

func (ActivityLevel) TableName() string {
	return "activity_levels"
}
