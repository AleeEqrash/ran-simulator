// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package model

// MapLayout represents information required for geo-map visualizations
type MapLayout struct {
	Center         Coordinate `yaml:"center"`
	Zoom           float32    `yaml:"zoom"`
	LocationsScale float32    `yaml:"locationsScale"`
	FadeMap        bool       `yaml:"fade"`
	ShowRoutes     bool       `yaml:"showRoutes"`
	ShowPower      bool       `yaml:"showPower"`
}