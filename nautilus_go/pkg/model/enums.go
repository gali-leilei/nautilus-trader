// -------------------------------------------------------------------------------------------------
//  Copyright (C) 2015-2026 Nautech Systems Pty Ltd. All rights reserved.
//  https://nautechsystems.io
//
//  Licensed under the GNU Lesser General Public License Version 3.0 (the "License");
//  You may not use this file except in compliance with the License.
//  You may obtain a copy of the License at https://www.gnu.org/licenses/lgpl-3.0.en.html
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
// -------------------------------------------------------------------------------------------------

package model

import (
	"fmt"
	"strings"
)

// BarAggregation represents how raw data is aggregated into bars.
type BarAggregation int

const (
	BarAggregationTick            BarAggregation = 1
	BarAggregationTickImbalance   BarAggregation = 2
	BarAggregationTickRuns        BarAggregation = 3
	BarAggregationVolume          BarAggregation = 4
	BarAggregationVolumeImbalance BarAggregation = 5
	BarAggregationVolumeRuns      BarAggregation = 6
	BarAggregationValue           BarAggregation = 7
	BarAggregationValueImbalance  BarAggregation = 8
	BarAggregationValueRuns       BarAggregation = 9
	BarAggregationMillisecond     BarAggregation = 10
	BarAggregationSecond          BarAggregation = 11
	BarAggregationMinute          BarAggregation = 12
	BarAggregationHour            BarAggregation = 13
	BarAggregationDay             BarAggregation = 14
	BarAggregationWeek            BarAggregation = 15
	BarAggregationMonth           BarAggregation = 16
	BarAggregationYear            BarAggregation = 17
)

var barAggregationNames = map[BarAggregation]string{
	BarAggregationTick:            "TICK",
	BarAggregationTickImbalance:   "TICK_IMBALANCE",
	BarAggregationTickRuns:        "TICK_RUNS",
	BarAggregationVolume:          "VOLUME",
	BarAggregationVolumeImbalance: "VOLUME_IMBALANCE",
	BarAggregationVolumeRuns:      "VOLUME_RUNS",
	BarAggregationValue:           "VALUE",
	BarAggregationValueImbalance:  "VALUE_IMBALANCE",
	BarAggregationValueRuns:       "VALUE_RUNS",
	BarAggregationMillisecond:     "MILLISECOND",
	BarAggregationSecond:          "SECOND",
	BarAggregationMinute:          "MINUTE",
	BarAggregationHour:            "HOUR",
	BarAggregationDay:             "DAY",
	BarAggregationWeek:            "WEEK",
	BarAggregationMonth:           "MONTH",
	BarAggregationYear:            "YEAR",
}

var barAggregationValues = func() map[string]BarAggregation {
	m := make(map[string]BarAggregation, len(barAggregationNames))
	for k, v := range barAggregationNames {
		m[v] = k
	}
	return m
}()

func (b BarAggregation) String() string {
	if s, ok := barAggregationNames[b]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(b))
}

// ParseBarAggregation parses a string into a BarAggregation.
func ParseBarAggregation(s string) (BarAggregation, error) {
	if v, ok := barAggregationValues[strings.ToUpper(s)]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid BarAggregation: %q", s)
}

// PriceType represents the type of price for a bar.
type PriceType int

const (
	PriceTypeBid  PriceType = 1
	PriceTypeAsk  PriceType = 2
	PriceTypeMid  PriceType = 3
	PriceTypeLast PriceType = 4
	PriceTypeMark PriceType = 5
)

var priceTypeNames = map[PriceType]string{
	PriceTypeBid:  "BID",
	PriceTypeAsk:  "ASK",
	PriceTypeMid:  "MID",
	PriceTypeLast: "LAST",
	PriceTypeMark: "MARK",
}

var priceTypeValues = func() map[string]PriceType {
	m := make(map[string]PriceType, len(priceTypeNames))
	for k, v := range priceTypeNames {
		m[v] = k
	}
	return m
}()

func (p PriceType) String() string {
	if s, ok := priceTypeNames[p]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(p))
}

// ParsePriceType parses a string into a PriceType.
func ParsePriceType(s string) (PriceType, error) {
	if v, ok := priceTypeValues[strings.ToUpper(s)]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid PriceType: %q", s)
}

// AggregationSource represents the source of bar aggregation.
type AggregationSource int

const (
	AggregationSourceExternal AggregationSource = 1
	AggregationSourceInternal AggregationSource = 2
)

var aggregationSourceNames = map[AggregationSource]string{
	AggregationSourceExternal: "EXTERNAL",
	AggregationSourceInternal: "INTERNAL",
}

var aggregationSourceValues = func() map[string]AggregationSource {
	m := make(map[string]AggregationSource, len(aggregationSourceNames))
	for k, v := range aggregationSourceNames {
		m[v] = k
	}
	return m
}()

func (a AggregationSource) String() string {
	if s, ok := aggregationSourceNames[a]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(a))
}

// ParseAggregationSource parses a string into an AggregationSource.
func ParseAggregationSource(s string) (AggregationSource, error) {
	if v, ok := aggregationSourceValues[strings.ToUpper(s)]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid AggregationSource: %q", s)
}

// OmsType represents the order management system type.
type OmsType int

const (
	OmsTypeNetting OmsType = 1
	OmsTypeHedging OmsType = 2
)

func (o OmsType) String() string {
	switch o {
	case OmsTypeNetting:
		return "NETTING"
	case OmsTypeHedging:
		return "HEDGING"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(o))
	}
}

// AccountType represents the type of trading account.
type AccountType int

const (
	AccountTypeCash   AccountType = 1
	AccountTypeMargin AccountType = 2
)

func (a AccountType) String() string {
	switch a {
	case AccountTypeCash:
		return "CASH"
	case AccountTypeMargin:
		return "MARGIN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(a))
	}
}

// AssetClass represents the asset class of an instrument.
type AssetClass int

const (
	AssetClassFX          AssetClass = 1
	AssetClassEquity      AssetClass = 2
	AssetClassCommodity   AssetClass = 3
	AssetClassDebt        AssetClass = 4
	AssetClassIndex       AssetClass = 5
	AssetClassCryptoCurr  AssetClass = 6
	AssetClassAlternative AssetClass = 7
)

func (a AssetClass) String() string {
	switch a {
	case AssetClassFX:
		return "FX"
	case AssetClassEquity:
		return "EQUITY"
	case AssetClassCommodity:
		return "COMMODITY"
	case AssetClassDebt:
		return "DEBT"
	case AssetClassIndex:
		return "INDEX"
	case AssetClassCryptoCurr:
		return "CRYPTOCURRENCY"
	case AssetClassAlternative:
		return "ALTERNATIVE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(a))
	}
}
