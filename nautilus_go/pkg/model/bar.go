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
	"strconv"
	"strings"
)

// BarSpecification defines bar aggregation parameters.
type BarSpecification struct {
	Step        uint32
	Aggregation BarAggregation
	PriceType   PriceType
}

func (s BarSpecification) String() string {
	return fmt.Sprintf(
		"%d-%s-%s",
		s.Step,
		s.Aggregation,
		s.PriceType,
	)
}

// BarType fully qualifies a bar stream.
type BarType struct {
	InstrumentId      InstrumentId
	Spec              BarSpecification
	AggregationSource AggregationSource
}

func (bt BarType) String() string {
	return fmt.Sprintf(
		"%s-%s-%s",
		bt.InstrumentId,
		bt.Spec,
		bt.AggregationSource,
	)
}

// Topic returns the message bus topic for this bar type.
func (bt BarType) Topic() string {
	return "data.bars." + bt.String()
}

// ParseBarType parses a string like "6EH4.XCME-1-MINUTE-LAST-EXTERNAL".
// Uses reverse-split on '-' to extract 5 components, since the instrument
// ID itself may contain hyphens.
func ParseBarType(s string) (BarType, error) {
	// Split off composite part (after @) — not needed for this example
	parts := strings.SplitN(s, "@", 2)
	standard := parts[0]

	// Reverse-split: find last 4 hyphens from the right
	tokens, e := rsplitN(standard, '-', 5)
	if e != nil {
		return BarType{}, fmt.Errorf("invalid BarType %q: %w", s, e)
	}

	instrumentId, e := ParseInstrumentId(tokens[0])
	if e != nil {
		return BarType{}, fmt.Errorf("invalid BarType %q: %w", s, e)
	}

	step, e := strconv.ParseUint(tokens[1], 10, 32)
	if e != nil {
		return BarType{}, fmt.Errorf(
			"invalid BarType step %q: %w", tokens[1], e,
		)
	}

	aggregation, e := ParseBarAggregation(tokens[2])
	if e != nil {
		return BarType{}, fmt.Errorf("invalid BarType %q: %w", s, e)
	}

	priceType, e := ParsePriceType(tokens[3])
	if e != nil {
		return BarType{}, fmt.Errorf("invalid BarType %q: %w", s, e)
	}

	aggSource, e := ParseAggregationSource(tokens[4])
	if e != nil {
		return BarType{}, fmt.Errorf("invalid BarType %q: %w", s, e)
	}

	return BarType{
		InstrumentId: instrumentId,
		Spec: BarSpecification{
			Step:        uint32(step),
			Aggregation: aggregation,
			PriceType:   priceType,
		},
		AggregationSource: aggSource,
	}, nil
}

// rsplitN splits s into n parts by delimiter, splitting from the right.
// Returns the parts in left-to-right order.
func rsplitN(s string, sep byte, n int) ([]string, error) {
	parts := make([]string, n)
	remaining := s
	for i := n - 1; i > 0; i-- {
		idx := strings.LastIndexByte(remaining, sep)
		if idx < 0 {
			return nil, fmt.Errorf(
				"expected %d parts separated by '%c', found fewer",
				n, sep,
			)
		}
		parts[i] = remaining[idx+1:]
		remaining = remaining[:idx]
	}
	parts[0] = remaining
	return parts, nil
}

// Bar represents an OHLCV bar.
type Bar struct {
	BarType BarType
	Open    Price
	High    Price
	Low     Price
	Close   Price
	Volume  Quantity
	TsEvent UnixNanos
	TsInit  UnixNanos
}

// IsSinglePrice returns true if all OHLC prices are equal.
func (b Bar) IsSinglePrice() bool {
	return b.Open.Raw == b.High.Raw &&
		b.High.Raw == b.Low.Raw &&
		b.Low.Raw == b.Close.Raw
}

func (b Bar) String() string {
	return fmt.Sprintf(
		"%s %s %s %s %s %s %d %d",
		b.BarType,
		b.Open,
		b.High,
		b.Low,
		b.Close,
		b.Volume,
		b.TsEvent,
		b.TsInit,
	)
}
