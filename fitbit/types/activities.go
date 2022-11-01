// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /activities.json

type ActivityCatalog struct {
	Categories []Category `json:"categories"`
}

type ActivityDescription struct {
	AccessLevel    string          `json:"accessLevel"`
	ActivityLevels []ActivityLevel `json:"activityLevels"`
	HasSpeed       bool            `json:"hasSpeed"`
	ID             int64           `json:"id"`
	Mets           float64         `json:"mets"`
	Name           string          `json:"name"`
}

type Category struct {
	Activities    []ActivityDescription `json:"activities"`
	ID            int64                 `json:"id"`
	Name          string                `json:"name"`
	SubCategories []SubCategory         `json:"subCategories"`
}

type SubCategory struct {
	Activities []ActivityDescription `json:"activities"`
	ID         int64                 `json:"id"`
	Name       string                `json:"name"`
}

type ActivityLevel struct {
	ID          int64   `json:"id"`
	MaxSpeedMPH float64 `json:"maxSpeedMPH"`
	Mets        int64   `json:"mets"`
	MinSpeedMPH int64   `json:"minSpeedMPH"`
	Name        string  `json:"name"`
}

// /activities/%s.json

type SingleActivity struct {
	Activity ActivityDescription `json:"activity"`
}
