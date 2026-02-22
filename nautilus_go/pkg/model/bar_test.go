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

func TestParseBarType(t *testing.T) {
	input := "6EH4.XCME-1-MINUTE-LAST-EXTERNAL"
	bt, e := ParseBarType(input)
	if e != nil {
		t.Fatalf("ParseBarType(%q) returned error: %v", input, e)
	}

	if bt.InstrumentId.Symbol.Value != "6EH4" {
		t.Errorf("symbol: expected 6EH4, found %s", bt.InstrumentId.Symbol.Value)
	}
	if bt.InstrumentId.Venue.Name != "XCME" {
		t.Errorf("venue: expected XCME, found %s", bt.InstrumentId.Venue.Name)
	}
	if bt.Spec.Step != 1 {
		t.Errorf("step: expected 1, found %d", bt.Spec.Step)
	}
	if bt.Spec.Aggregation != BarAggregationMinute {
		t.Errorf("aggregation: expected MINUTE, found %s", bt.Spec.Aggregation)
	}
	if bt.Spec.PriceType != PriceTypeLast {
		t.Errorf("price_type: expected LAST, found %s", bt.Spec.PriceType)
	}
	if bt.AggregationSource != AggregationSourceExternal {
		t.Errorf("source: expected EXTERNAL, found %s", bt.AggregationSource)
	}
}

func TestParseBarTypeRoundtrip(t *testing.T) {
	input := "6EH4.XCME-1-MINUTE-LAST-EXTERNAL"
	bt, e := ParseBarType(input)
	if e != nil {
		t.Fatalf("ParseBarType returned error: %v", e)
	}
	result := bt.String()
	if result != input {
		t.Errorf("roundtrip: expected %q, found %q", input, result)
	}
}

func TestParseBarTypeWithHyphenatedSymbol(t *testing.T) {
	input := "EUR-USD.FOREX-5-MINUTE-BID-INTERNAL"
	bt, e := ParseBarType(input)
	if e != nil {
		t.Fatalf("ParseBarType(%q) returned error: %v", input, e)
	}
	if bt.InstrumentId.Symbol.Value != "EUR-USD" {
		t.Errorf("symbol: expected EUR-USD, found %s", bt.InstrumentId.Symbol.Value)
	}
	if bt.Spec.Step != 5 {
		t.Errorf("step: expected 5, found %d", bt.Spec.Step)
	}
}

func TestParseBarTypeInvalid(t *testing.T) {
	cases := []string{
		"",
		"INVALID",
		"SYM.VENUE-1-MINUTE",
		"SYM.VENUE-1-MINUTE-LAST",
	}
	for _, input := range cases {
		_, e := ParseBarType(input)
		if e == nil {
			t.Errorf("ParseBarType(%q) expected error, found nil", input)
		}
	}
}

func TestBarIsSinglePriceTrue(t *testing.T) {
	bt, _ := ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	bar := Bar{
		BarType: bt,
		Open:    NewPrice(1.10000, 5),
		High:    NewPrice(1.10000, 5),
		Low:     NewPrice(1.10000, 5),
		Close:   NewPrice(1.10000, 5),
		Volume:  NewQuantity(100, 0),
	}
	if !bar.IsSinglePrice() {
		t.Error("expected IsSinglePrice() = true for equal OHLC")
	}
}

func TestBarIsSinglePriceFalse(t *testing.T) {
	bt, _ := ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	bar := Bar{
		BarType: bt,
		Open:    NewPrice(1.10000, 5),
		High:    NewPrice(1.10010, 5),
		Low:     NewPrice(1.09990, 5),
		Close:   NewPrice(1.10005, 5),
		Volume:  NewQuantity(100, 0),
	}
	if bar.IsSinglePrice() {
		t.Error("expected IsSinglePrice() = false for different OHLC")
	}
}

func TestBarTypeTopic(t *testing.T) {
	bt, _ := ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	expected := "data.bars.6EH4.XCME-1-MINUTE-LAST-EXTERNAL"
	if bt.Topic() != expected {
		t.Errorf("topic: expected %q, found %q", expected, bt.Topic())
	}
}
