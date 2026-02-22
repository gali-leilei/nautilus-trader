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
	"math"
	"testing"
)

func TestNewPriceBasic(t *testing.T) {
	p := NewPrice(1.2345, 4)
	if p.Precision != 4 {
		t.Errorf("precision: expected 4, found %d", p.Precision)
	}
	// raw should be round(1.2345 * 10^4) * 10^(9-4) = 12345 * 100000 = 1234500000
	expectedRaw := int64(1_234_500_000)
	if p.Raw != expectedRaw {
		t.Errorf("raw: expected %d, found %d", expectedRaw, p.Raw)
	}
}

func TestPriceAsFloat64(t *testing.T) {
	p := NewPrice(1.10780, 5)
	got := p.AsFloat64()
	if math.Abs(got-1.10780) > 1e-10 {
		t.Errorf("AsFloat64: expected 1.10780, found %v", got)
	}
}

func TestPriceString(t *testing.T) {
	p := NewPrice(1.10780, 5)
	s := p.String()
	if s != "1.10780" {
		t.Errorf("String: expected 1.10780, found %s", s)
	}
}

func TestPriceZero(t *testing.T) {
	p := NewPrice(0, 2)
	if p.Raw != 0 {
		t.Errorf("zero price raw: expected 0, found %d", p.Raw)
	}
	if p.String() != "0.00" {
		t.Errorf("zero price string: expected 0.00, found %s", p.String())
	}
}

func TestNewQuantityBasic(t *testing.T) {
	q := NewQuantity(205, 0)
	// raw = round(205 * 10^0) * 10^(9-0) = 205 * 1_000_000_000 = 205_000_000_000
	expectedRaw := uint64(205_000_000_000)
	if q.Raw != expectedRaw {
		t.Errorf("raw: expected %d, found %d", expectedRaw, q.Raw)
	}
}

func TestQuantityAsFloat64(t *testing.T) {
	q := NewQuantity(205, 0)
	got := q.AsFloat64()
	if math.Abs(got-205.0) > 1e-10 {
		t.Errorf("AsFloat64: expected 205, found %v", got)
	}
}

func TestQuantityString(t *testing.T) {
	q := NewQuantity(205, 0)
	if q.String() != "205" {
		t.Errorf("String: expected 205, found %s", q.String())
	}
}
