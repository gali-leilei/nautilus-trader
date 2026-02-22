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

import "fmt"

// CurrencyType represents fiat vs crypto.
type CurrencyType int

const (
	CurrencyTypeFiat   CurrencyType = 1
	CurrencyTypeCrypto CurrencyType = 2
)

// Currency represents a monetary currency.
type Currency struct {
	Code         string
	Precision    uint8
	ISO4217      uint16
	Name         string
	CurrencyType CurrencyType
}

func (c Currency) String() string {
	return c.Code
}

// Pre-defined currencies
var USD = Currency{
	Code:         "USD",
	Precision:    2,
	ISO4217:      840,
	Name:         "United States dollar",
	CurrencyType: CurrencyTypeFiat,
}

// Money represents a monetary amount in a specific currency.
type Money struct {
	Raw      int64
	Currency Currency
}

// NewMoney creates a Money from a float64 amount and currency.
func NewMoney(amount float64, currency Currency) Money {
	return Money{
		Raw:      f64ToFixedI64(amount, currency.Precision),
		Currency: currency,
	}
}

// AsFloat64 returns the money amount as a float64.
func (m Money) AsFloat64() float64 {
	return fixedI64ToF64(m.Raw)
}

func (m Money) String() string {
	return fmt.Sprintf(
		"%.*f %s",
		m.Currency.Precision,
		m.AsFloat64(),
		m.Currency.Code,
	)
}
