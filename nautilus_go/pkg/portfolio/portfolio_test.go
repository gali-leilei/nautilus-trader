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

func makeInstrument1() model.InstrumentId {
	return model.NewInstrumentId(
		model.Symbol{Value: "6EH4"},
		model.Venue{Name: "XCME"},
	)
}

func makeInstrument2() model.InstrumentId {
	return model.NewInstrumentId(
		model.Symbol{Value: "ESH4"},
		model.Venue{Name: "XCME"},
	)
}

func makePortfolioFill(
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

func TestEmptyPortfolioIsFlat(t *testing.T) {
	p := NewPortfolio()
	id := makeInstrument1()
	if !p.IsFlat(id) {
		t.Error("empty portfolio should be flat for any instrument")
	}
	if p.IsNetLong(id) {
		t.Error("empty portfolio should not be net long")
	}
	if p.IsNetShort(id) {
		t.Error("empty portfolio should not be net short")
	}
	if p.Position(id) != nil {
		t.Error("empty portfolio Position() should return nil")
	}
}

func TestPortfolioAfterBuy(t *testing.T) {
	p := NewPortfolio()
	id := makeInstrument1()
	p.ApplyFill(makePortfolioFill(model.OrderSideBuy, 10, id))

	if p.IsFlat(id) {
		t.Error("should not be flat after buy")
	}
	if !p.IsNetLong(id) {
		t.Error("should be net long after buy")
	}
	if p.IsNetShort(id) {
		t.Error("should not be net short after buy")
	}
	pos := p.Position(id)
	if pos == nil {
		t.Fatal("Position() should not be nil after buy")
	}
	if pos.SignedQty != 10 {
		t.Errorf("SignedQty = %v, expected 10", pos.SignedQty)
	}
}

func TestPortfolioAfterSell(t *testing.T) {
	p := NewPortfolio()
	id := makeInstrument1()
	p.ApplyFill(makePortfolioFill(model.OrderSideSell, 5, id))

	if !p.IsNetShort(id) {
		t.Error("should be net short after sell")
	}
}

func TestPortfolioRoundtrip(t *testing.T) {
	p := NewPortfolio()
	id := makeInstrument1()
	p.ApplyFill(makePortfolioFill(model.OrderSideBuy, 10, id))
	p.ApplyFill(makePortfolioFill(model.OrderSideSell, 10, id))

	if !p.IsFlat(id) {
		t.Error("should be flat after roundtrip")
	}
}

func TestPortfolioMultipleInstruments(t *testing.T) {
	p := NewPortfolio()
	id1 := makeInstrument1()
	id2 := makeInstrument2()

	p.ApplyFill(makePortfolioFill(model.OrderSideBuy, 10, id1))
	p.ApplyFill(makePortfolioFill(model.OrderSideSell, 5, id2))

	if !p.IsNetLong(id1) {
		t.Error("instrument1 should be net long")
	}
	if !p.IsNetShort(id2) {
		t.Error("instrument2 should be net short")
	}

	// Instruments are independent
	if p.IsNetShort(id1) {
		t.Error("instrument1 should not be net short")
	}
	if p.IsNetLong(id2) {
		t.Error("instrument2 should not be net long")
	}
}
