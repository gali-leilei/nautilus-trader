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

import "time"

// Instrument defines the interface for a tradeable instrument.
type Instrument interface {
	ID() InstrumentId
	PricePrecision() uint8
	SizePrecision() uint8
	QuoteCurrency() Currency
}

// FuturesContract represents a futures instrument.
type FuturesContract struct {
	InstrumentId   InstrumentId
	RawSymbol      Symbol
	AssetClass     AssetClass
	QuoteCurr      Currency
	Underlying     string
	ActivationNs   UnixNanos
	ExpirationNs   UnixNanos
	PricePrec      uint8
	SizePrec       uint8
	PriceIncrement Price
	SizeIncrement  Quantity
	Multiplier     Quantity
	LotSize        Quantity
	Expiry         time.Time
}

func (f *FuturesContract) ID() InstrumentId      { return f.InstrumentId }
func (f *FuturesContract) PricePrecision() uint8  { return f.PricePrec }
func (f *FuturesContract) SizePrecision() uint8   { return f.SizePrec }
func (f *FuturesContract) QuoteCurrency() Currency { return f.QuoteCurr }
