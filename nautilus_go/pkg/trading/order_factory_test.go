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

package trading

import (
	"testing"

	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/model"
)

func makeTestOrderFactory() *OrderFactory {
	clock := common.NewTestClock()
	clock.SetTime(1000)
	return NewOrderFactory(
		model.TraderId{Value: "TRADER-001"},
		model.StrategyId{Value: "EMACross-001"},
		clock,
	)
}

func makeTestInstrumentId() model.InstrumentId {
	return model.NewInstrumentId(
		model.Symbol{Value: "6EH4"},
		model.Venue{Name: "XCME"},
	)
}

func TestOrderFactoryUniqueIds(t *testing.T) {
	factory := makeTestOrderFactory()
	id := makeTestInstrumentId()

	order1 := factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(1, 0), model.TimeInForceGTC,
	)
	order2 := factory.Market(
		id, model.OrderSideSell, model.NewQuantity(2, 0), model.TimeInForceGTC,
	)

	if order1.ClientOrderId.Value == order2.ClientOrderId.Value {
		t.Errorf(
			"expected unique IDs, both are %q",
			order1.ClientOrderId.Value,
		)
	}
}

func TestOrderFactoryIdFormat(t *testing.T) {
	factory := makeTestOrderFactory()
	id := makeTestInstrumentId()

	order := factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(1, 0), model.TimeInForceGTC,
	)

	expected := "O-TRADER-001-EMACross-001-1"
	if order.ClientOrderId.Value != expected {
		t.Errorf(
			"ClientOrderId = %q, expected %q",
			order.ClientOrderId.Value, expected,
		)
	}
}

func TestOrderFactoryFields(t *testing.T) {
	factory := makeTestOrderFactory()
	id := makeTestInstrumentId()

	order := factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(10, 0), model.TimeInForceGTC,
	)

	if order.Side != model.OrderSideBuy {
		t.Errorf("Side = %v, expected BUY", order.Side)
	}
	if order.Quantity.AsFloat64() != 10 {
		t.Errorf("Quantity = %v, expected 10", order.Quantity.AsFloat64())
	}
	if order.Status != model.OrderStatusInitialized {
		t.Errorf("Status = %v, expected INITIALIZED", order.Status)
	}
	if order.TimeInForce != model.TimeInForceGTC {
		t.Errorf("TimeInForce = %v, expected GTC", order.TimeInForce)
	}
	if order.TsInit != 1000 {
		t.Errorf("TsInit = %d, expected 1000", order.TsInit)
	}
	if order.InstrumentId.String() != "6EH4.XCME" {
		t.Errorf("InstrumentId = %s, expected 6EH4.XCME", order.InstrumentId)
	}
}

func TestOrderFactoryReset(t *testing.T) {
	factory := makeTestOrderFactory()
	id := makeTestInstrumentId()

	factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(1, 0), model.TimeInForceGTC,
	)
	factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(1, 0), model.TimeInForceGTC,
	)

	factory.Reset()

	order := factory.Market(
		id, model.OrderSideBuy, model.NewQuantity(1, 0), model.TimeInForceGTC,
	)
	expected := "O-TRADER-001-EMACross-001-1"
	if order.ClientOrderId.Value != expected {
		t.Errorf(
			"after Reset: ClientOrderId = %q, expected %q",
			order.ClientOrderId.Value, expected,
		)
	}
}
