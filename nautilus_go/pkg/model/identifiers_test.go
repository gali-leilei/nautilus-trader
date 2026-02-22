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

func TestParseInstrumentId(t *testing.T) {
	id, e := ParseInstrumentId("6EH4.XCME")
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if id.Symbol.Value != "6EH4" {
		t.Errorf("symbol: expected 6EH4, found %s", id.Symbol.Value)
	}
	if id.Venue.Name != "XCME" {
		t.Errorf("venue: expected XCME, found %s", id.Venue.Name)
	}
}

func TestParseInstrumentIdWithDots(t *testing.T) {
	id, e := ParseInstrumentId("BTC.USD.BINANCE")
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if id.Symbol.Value != "BTC.USD" {
		t.Errorf("symbol: expected BTC.USD, found %s", id.Symbol.Value)
	}
	if id.Venue.Name != "BINANCE" {
		t.Errorf("venue: expected BINANCE, found %s", id.Venue.Name)
	}
}

func TestParseInstrumentIdRoundtrip(t *testing.T) {
	input := "6EH4.XCME"
	id, _ := ParseInstrumentId(input)
	if id.String() != input {
		t.Errorf("roundtrip: expected %q, found %q", input, id.String())
	}
}

func TestParseInstrumentIdInvalid(t *testing.T) {
	cases := []string{"", "NODOT", ".VENUE", "SYM."}
	for _, input := range cases {
		_, e := ParseInstrumentId(input)
		if e == nil {
			t.Errorf("ParseInstrumentId(%q) expected error, found nil", input)
		}
	}
}
