// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"github.com/galeone/fitbit/types"
)

type Category struct {
	types.Category
	ID int64 `igor:"primary_key"`
}

func (Category) TableName() string {
	return "categories"
}

type SubCategory struct {
	types.SubCategory
	ID         int64    `igor:"primary_key"`
	Category   Category `sql:"-"`
	CategoryID int64
}

func (SubCategory) TableName() string {
	return "subcategories"
}

type ActivityDescription struct {
	types.ActivityDescription
	ID            int64       `igor:"primary_key"`
	SubCategory   SubCategory `sql:"-"`
	SubCategoryID int64
	Category      Category `sql:"-"`
	CategoryID    int64
}

func (ActivityDescription) TableName() string {
	return "activities_descriptions"
}

type ActivityLevel struct {
	types.ActivityLevel
	ID                    int64               `igor:"primary_key"`
	ActivityDescription   ActivityDescription `sql:"-"`
	ActivityDescriptionID int64
}

func (ActivityLevel) TableName() string {
	return "activity_levels"
}
