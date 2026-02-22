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

package backtest

import (
	"log/slog"
	"testing"

	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/model"
	"github.com/nautechsystems/nautilus_go/pkg/portfolio"
)

func makeTestExchange() *SimulatedExchange {
	logger := common.NewLogger(slog.LevelDebug)
	port := portfolio.NewPortfolio()
	return NewSimulatedExchange(
		model.Venue{Name: "XCME"},
		port,
		logger,
	)
}

func makeTestBar() model.Bar {
	bt, _ := model.ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	return model.Bar{
		BarType: bt,
		Open:    model.NewPrice(1.10000, 5),
		High:    model.NewPrice(1.10100, 5),
		Low:     model.NewPrice(1.09900, 5),
		Close:   model.NewPrice(1.10050, 5),
		Volume:  model.NewQuantity(100, 0),
		TsEvent: 1000,
		TsInit:  1000,
	}
}

func makeTestOrder(side model.OrderSide) model.MarketOrder {
	return model.MarketOrder{
		ClientOrderId: model.ClientOrderId{Value: "O-001"},
		InstrumentId: model.NewInstrumentId(
			model.Symbol{Value: "6EH4"},
			model.Venue{Name: "XCME"},
		),
		Side:        side,
		Quantity:    model.NewQuantity(1, 0),
		TimeInForce: model.TimeInForceGTC,
		Status:      model.OrderStatusInitialized,
		TsInit:      1000,
	}
}

func TestExchangeSubmitAndProcess(t *testing.T) {
	ex := makeTestExchange()
	bar := makeTestBar()
	order := makeTestOrder(model.OrderSideBuy)

	var received []model.OrderFilled
	ex.OnFill(func(fill model.OrderFilled) {
		received = append(received, fill)
	})

	ex.ProcessBar(bar)
	ex.SubmitOrder(order)
	ex.ProcessOrders()

	if len(received) != 1 {
		t.Fatalf("expected 1 fill, found %d", len(received))
	}

	fill := received[0]
	if fill.ClientOrderId.Value != "O-001" {
		t.Errorf("fill ClientOrderId = %q, expected %q", fill.ClientOrderId.Value, "O-001")
	}
	if fill.Side != model.OrderSideBuy {
		t.Errorf("fill Side = %v, expected BUY", fill.Side)
	}
	if fill.FillPrice.AsFloat64() != bar.Close.AsFloat64() {
		t.Errorf("fill price = %v, expected %v", fill.FillPrice.AsFloat64(), bar.Close.AsFloat64())
	}
}

func TestExchangeNoBarNoFills(t *testing.T) {
	ex := makeTestExchange()
	order := makeTestOrder(model.OrderSideBuy)

	var received []model.OrderFilled
	ex.OnFill(func(fill model.OrderFilled) {
		received = append(received, fill)
	})

	ex.SubmitOrder(order)
	ex.ProcessOrders()

	if len(received) != 0 {
		t.Errorf("expected 0 fills without bar, found %d", len(received))
	}
}

func TestExchangeProcessOrdersClearsPending(t *testing.T) {
	ex := makeTestExchange()
	bar := makeTestBar()
	order := makeTestOrder(model.OrderSideBuy)

	ex.ProcessBar(bar)
	ex.SubmitOrder(order)
	ex.ProcessOrders()

	// Second process should produce no fills
	var received []model.OrderFilled
	ex.OnFill(func(fill model.OrderFilled) {
		received = append(received, fill)
	})
	ex.ProcessOrders()

	if len(received) != 0 {
		t.Errorf("expected 0 fills after clearing, found %d", len(received))
	}
}

func TestExchangeFillUpdatesPortfolio(t *testing.T) {
	ex := makeTestExchange()
	bar := makeTestBar()
	order := makeTestOrder(model.OrderSideBuy)
	id := order.InstrumentId

	ex.ProcessBar(bar)
	ex.SubmitOrder(order)
	ex.ProcessOrders()

	if !ex.Portfolio.IsNetLong(id) {
		t.Error("portfolio should be net long after buy fill")
	}
}

func TestExchangeCancelAllOrders(t *testing.T) {
	ex := makeTestExchange()
	bar := makeTestBar()

	order1 := makeTestOrder(model.OrderSideBuy)
	order2 := model.MarketOrder{
		ClientOrderId: model.ClientOrderId{Value: "O-002"},
		InstrumentId: model.NewInstrumentId(
			model.Symbol{Value: "ESH4"},
			model.Venue{Name: "XCME"},
		),
		Side:        model.OrderSideSell,
		Quantity:    model.NewQuantity(1, 0),
		TimeInForce: model.TimeInForceGTC,
		Status:      model.OrderStatusInitialized,
	}

	ex.SubmitOrder(order1)
	ex.SubmitOrder(order2)

	// Cancel only 6EH4 orders
	ex.CancelAllOrders(order1.InstrumentId)

	var received []model.OrderFilled
	ex.OnFill(func(fill model.OrderFilled) {
		received = append(received, fill)
	})

	ex.ProcessBar(bar)
	ex.ProcessOrders()

	if len(received) != 1 {
		t.Fatalf("expected 1 fill (ESH4 only), found %d", len(received))
	}
	if received[0].InstrumentId.String() != "ESH4.XCME" {
		t.Errorf("remaining fill instrument = %s, expected ESH4.XCME", received[0].InstrumentId)
	}
}

func TestExchangeMultipleHandlers(t *testing.T) {
	ex := makeTestExchange()
	bar := makeTestBar()
	order := makeTestOrder(model.OrderSideBuy)

	count1 := 0
	count2 := 0
	ex.OnFill(func(_ model.OrderFilled) { count1++ })
	ex.OnFill(func(_ model.OrderFilled) { count2++ })

	ex.ProcessBar(bar)
	ex.SubmitOrder(order)
	ex.ProcessOrders()

	if count1 != 1 {
		t.Errorf("handler1 called %d times, expected 1", count1)
	}
	if count2 != 1 {
		t.Errorf("handler2 called %d times, expected 1", count2)
	}
}
