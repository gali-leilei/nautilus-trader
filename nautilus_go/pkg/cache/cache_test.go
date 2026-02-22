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

package cache

import (
	"testing"

	"github.com/nautechsystems/nautilus_go/pkg/model"
)

func makeTestBarType() model.BarType {
	bt, _ := model.ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	return bt
}

func TestCacheBarEmpty(t *testing.T) {
	c := NewCache()
	bt := makeTestBarType()
	if c.Bar(bt) != nil {
		t.Error("expected nil for empty cache")
	}
	if c.BarCount(bt) != 0 {
		t.Errorf("expected 0 bars, found %d", c.BarCount(bt))
	}
}

func TestCacheAddBar(t *testing.T) {
	c := NewCache()
	bt := makeTestBarType()
	bar := model.Bar{
		BarType: bt,
		Open:    model.NewPrice(1.10, 2),
		High:    model.NewPrice(1.11, 2),
		Low:     model.NewPrice(1.09, 2),
		Close:   model.NewPrice(1.10, 2),
		Volume:  model.NewQuantity(100, 0),
		TsEvent: 1000,
		TsInit:  1000,
	}

	c.AddBar(bar)

	if c.BarCount(bt) != 1 {
		t.Errorf("expected 1 bar, found %d", c.BarCount(bt))
	}
	last := c.Bar(bt)
	if last == nil {
		t.Fatal("expected non-nil bar")
	}
	if last.TsInit != 1000 {
		t.Errorf("expected TsInit 1000, found %d", last.TsInit)
	}
}

func TestCacheInstrument(t *testing.T) {
	c := NewCache()
	inst := &model.FuturesContract{
		InstrumentId: model.NewInstrumentId(
			model.Symbol{Value: "6EH4"},
			model.Venue{Name: "XCME"},
		),
		PricePrec: 5,
		SizePrec:  0,
	}

	c.AddInstrument(inst)

	got := c.Instrument(inst.ID())
	if got == nil {
		t.Fatal("expected non-nil instrument")
	}
	if got.ID().String() != "6EH4.XCME" {
		t.Errorf("expected 6EH4.XCME, found %s", got.ID())
	}
}
