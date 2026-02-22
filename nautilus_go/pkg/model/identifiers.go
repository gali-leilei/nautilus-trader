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

// Symbol represents a trading symbol.
type Symbol struct {
	Value string
}

func (s Symbol) String() string {
	return s.Value
}

// Venue represents a trading venue.
type Venue struct {
	Name string
}

func (v Venue) String() string {
	return v.Name
}

// InstrumentId uniquely identifies a tradeable instrument.
type InstrumentId struct {
	Symbol Symbol
	Venue  Venue
}

func NewInstrumentId(symbol Symbol, venue Venue) InstrumentId {
	return InstrumentId{Symbol: symbol, Venue: venue}
}

// ParseInstrumentId parses "SYMBOL.VENUE" splitting on the last dot.
func ParseInstrumentId(s string) (InstrumentId, error) {
	idx := strings.LastIndex(s, ".")
	if idx < 0 || idx == 0 || idx == len(s)-1 {
		return InstrumentId{}, fmt.Errorf(
			"invalid InstrumentId format: %q, expected SYMBOL.VENUE", s,
		)
	}
	return InstrumentId{
		Symbol: Symbol{Value: s[:idx]},
		Venue:  Venue{Name: s[idx+1:]},
	}, nil
}

func (id InstrumentId) String() string {
	return id.Symbol.Value + "." + id.Venue.Name
}

// TraderId identifies a trader instance.
type TraderId struct {
	Value string
}

func (t TraderId) String() string {
	return t.Value
}
