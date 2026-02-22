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

package portfolio

import (
	"testing"

	"github.com/nautechsystems/nautilus_go/pkg/model"
)

func makeTestInstrumentId() model.InstrumentId {
	return model.NewInstrumentId(
		model.Symbol{Value: "6EH4"},
		model.Venue{Name: "XCME"},
	)
}

func makeFill(
	side model.OrderSide,
	qty float64,
	instrumentId model.InstrumentId,
) model.OrderFilled {
	return model.OrderFilled{
		ClientOrderId: model.ClientOrderId{Value: "O-001"},
		InstrumentId:  instrumentId,
		Side:          side,
		Quantity:      model.NewQuantity(qty, 0),
		FillPrice:     model.NewPrice(1.10000, 5),
		TsEvent:       1000,
		TsInit:        1000,
	}
}

func TestNewPositionIsFlat(t *testing.T) {
	pos := NewPosition(makeTestInstrumentId())
	if !pos.IsFlat() {
		t.Error("new position should be flat")
	}
	if pos.IsLong() {
		t.Error("new position should not be long")
	}
	if pos.IsShort() {
		t.Error("new position should not be short")
	}
	if pos.SignedQty != 0 {
		t.Errorf("new position SignedQty = %v, expected 0", pos.SignedQty)
	}
}

func TestBuyFillMakesLong(t *testing.T) {
	id := makeTestInstrumentId()
	pos := NewPosition(id)
	pos.ApplyFill(makeFill(model.OrderSideBuy, 10, id))

	if !pos.IsLong() {
		t.Error("expected long after buy fill")
	}
	if pos.SignedQty != 10 {
		t.Errorf("SignedQty = %v, expected 10", pos.SignedQty)
	}
	if pos.Quantity.AsFloat64() != 10 {
		t.Errorf("Quantity = %v, expected 10", pos.Quantity.AsFloat64())
	}
}

func TestSellFillMakesShort(t *testing.T) {
	id := makeTestInstrumentId()
	pos := NewPosition(id)
	pos.ApplyFill(makeFill(model.OrderSideSell, 5, id))

	if !pos.IsShort() {
		t.Error("expected short after sell fill")
	}
	if pos.SignedQty != -5 {
		t.Errorf("SignedQty = %v, expected -5", pos.SignedQty)
	}
}

func TestBuyThenSellSameQtyFlat(t *testing.T) {
	id := makeTestInstrumentId()
	pos := NewPosition(id)
	pos.ApplyFill(makeFill(model.OrderSideBuy, 10, id))
	pos.ApplyFill(makeFill(model.OrderSideSell, 10, id))

	if !pos.IsFlat() {
		t.Error("expected flat after buy+sell same qty")
	}
	if pos.SignedQty != 0 {
		t.Errorf("SignedQty = %v, expected 0", pos.SignedQty)
	}
}

func TestPartialClose(t *testing.T) {
	id := makeTestInstrumentId()
	pos := NewPosition(id)
	pos.ApplyFill(makeFill(model.OrderSideBuy, 10, id))
	pos.ApplyFill(makeFill(model.OrderSideSell, 5, id))

	if !pos.IsLong() {
		t.Error("expected long after partial close")
	}
	if pos.SignedQty != 5 {
		t.Errorf("SignedQty = %v, expected 5", pos.SignedQty)
	}
	if pos.Quantity.AsFloat64() != 5 {
		t.Errorf("Quantity = %v, expected 5", pos.Quantity.AsFloat64())
	}
}

func TestReversal(t *testing.T) {
	id := makeTestInstrumentId()
	pos := NewPosition(id)
	pos.ApplyFill(makeFill(model.OrderSideSell, 5, id))

	if !pos.IsShort() {
		t.Error("expected short after sell")
	}

	pos.ApplyFill(makeFill(model.OrderSideBuy, 10, id))

	if !pos.IsLong() {
		t.Error("expected long after reversal")
	}
	if pos.SignedQty != 5 {
		t.Errorf("SignedQty = %v, expected 5", pos.SignedQty)
	}
}
