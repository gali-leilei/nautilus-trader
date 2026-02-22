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
	"math"
)

// FixedPrecision is the number of decimal digits stored internally.
// Standard precision mode uses 9 digits with int64 backing.
const FixedPrecision uint8 = 9

// FixedScalar is 10^FixedPrecision, the internal scaling factor.
const FixedScalar float64 = 1_000_000_000.0

// pow10 returns 10^exp as int64.
func pow10(exp uint8) int64 {
	result := int64(1)
	for i := uint8(0); i < exp; i++ {
		result *= 10
	}
	return result
}

// f64ToFixedI64 converts a float64 value with the given user precision
// to an int64 fixed-point representation at FixedPrecision.
func f64ToFixedI64(value float64, precision uint8) int64 {
	pow1 := pow10(precision)
	pow2 := pow10(FixedPrecision - precision)
	rounded := int64(math.Round(value * float64(pow1)))
	return rounded * pow2
}

// fixedI64ToF64 converts an internal fixed-point int64 back to float64.
func fixedI64ToF64(value int64) float64 {
	return float64(value) / FixedScalar
}

// Price represents a fixed-point price value.
type Price struct {
	Raw       int64
	Precision uint8
}

// NewPrice creates a Price from a float64 with the given precision.
func NewPrice(value float64, precision uint8) Price {
	return Price{
		Raw:       f64ToFixedI64(value, precision),
		Precision: precision,
	}
}

// AsFloat64 returns the price as a float64.
func (p Price) AsFloat64() float64 {
	return fixedI64ToF64(p.Raw)
}

func (p Price) String() string {
	return fmt.Sprintf("%.*f", p.Precision, p.AsFloat64())
}

// Quantity represents a fixed-point quantity (non-negative).
type Quantity struct {
	Raw       uint64
	Precision uint8
}

// NewQuantity creates a Quantity from a float64 with the given precision.
func NewQuantity(value float64, precision uint8) Quantity {
	raw := f64ToFixedI64(value, precision)
	return Quantity{
		Raw:       uint64(raw),
		Precision: precision,
	}
}

// AsFloat64 returns the quantity as a float64.
func (q Quantity) AsFloat64() float64 {
	return float64(q.Raw) / FixedScalar
}

func (q Quantity) String() string {
	return fmt.Sprintf("%.*f", q.Precision, q.AsFloat64())
}

// UnixNanos represents a UNIX timestamp in nanoseconds.
type UnixNanos uint64
