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

import "testing"

func TestParseBarAggregation(t *testing.T) {
	tests := []struct {
		input    string
		expected BarAggregation
	}{
		{"MINUTE", BarAggregationMinute},
		{"HOUR", BarAggregationHour},
		{"DAY", BarAggregationDay},
		{"SECOND", BarAggregationSecond},
		{"TICK", BarAggregationTick},
	}
	for _, tc := range tests {
		got, e := ParseBarAggregation(tc.input)
		if e != nil {
			t.Errorf("ParseBarAggregation(%q) error: %v", tc.input, e)
		}
		if got != tc.expected {
			t.Errorf("ParseBarAggregation(%q) = %v, expected %v", tc.input, got, tc.expected)
		}
	}
}

func TestParseBarAggregationCaseInsensitive(t *testing.T) {
	got, e := ParseBarAggregation("minute")
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if got != BarAggregationMinute {
		t.Errorf("expected MINUTE, found %s", got)
	}
}

func TestParseBarAggregationInvalid(t *testing.T) {
	_, e := ParseBarAggregation("INVALID")
	if e == nil {
		t.Error("expected error for invalid aggregation")
	}
}

func TestParsePriceType(t *testing.T) {
	tests := []struct {
		input    string
		expected PriceType
	}{
		{"BID", PriceTypeBid},
		{"ASK", PriceTypeAsk},
		{"MID", PriceTypeMid},
		{"LAST", PriceTypeLast},
		{"MARK", PriceTypeMark},
	}
	for _, tc := range tests {
		got, e := ParsePriceType(tc.input)
		if e != nil {
			t.Errorf("ParsePriceType(%q) error: %v", tc.input, e)
		}
		if got != tc.expected {
			t.Errorf("ParsePriceType(%q) = %v, expected %v", tc.input, got, tc.expected)
		}
	}
}

func TestParseAggregationSource(t *testing.T) {
	got, e := ParseAggregationSource("EXTERNAL")
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if got != AggregationSourceExternal {
		t.Errorf("expected EXTERNAL, found %s", got)
	}

	got, e = ParseAggregationSource("INTERNAL")
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if got != AggregationSourceInternal {
		t.Errorf("expected INTERNAL, found %s", got)
	}
}
