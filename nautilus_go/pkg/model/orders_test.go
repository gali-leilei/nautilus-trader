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

import "testing"

func TestOrderSideString(t *testing.T) {
	tests := []struct {
		side     OrderSide
		expected string
	}{
		{OrderSideBuy, "BUY"},
		{OrderSideSell, "SELL"},
		{OrderSide(99), "UNKNOWN(99)"},
	}
	for _, tc := range tests {
		got := tc.side.String()
		if got != tc.expected {
			t.Errorf("OrderSide(%d).String() = %q, expected %q", int(tc.side), got, tc.expected)
		}
	}
}

func TestTimeInForceString(t *testing.T) {
	tests := []struct {
		tif      TimeInForce
		expected string
	}{
		{TimeInForceGTC, "GTC"},
		{TimeInForceIOC, "IOC"},
		{TimeInForceFOK, "FOK"},
		{TimeInForceDAY, "DAY"},
		{TimeInForce(99), "UNKNOWN(99)"},
	}
	for _, tc := range tests {
		got := tc.tif.String()
		if got != tc.expected {
			t.Errorf("TimeInForce(%d).String() = %q, expected %q", int(tc.tif), got, tc.expected)
		}
	}
}

func TestOrderStatusString(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusInitialized, "INITIALIZED"},
		{OrderStatusSubmitted, "SUBMITTED"},
		{OrderStatusAccepted, "ACCEPTED"},
		{OrderStatusFilled, "FILLED"},
		{OrderStatusCanceled, "CANCELED"},
		{OrderStatus(99), "UNKNOWN(99)"},
	}
	for _, tc := range tests {
		got := tc.status.String()
		if got != tc.expected {
			t.Errorf("OrderStatus(%d).String() = %q, expected %q", int(tc.status), got, tc.expected)
		}
	}
}

func TestPositionSideString(t *testing.T) {
	tests := []struct {
		side     PositionSide
		expected string
	}{
		{PositionSideFlat, "FLAT"},
		{PositionSideLong, "LONG"},
		{PositionSideShort, "SHORT"},
		{PositionSide(99), "UNKNOWN(99)"},
	}
	for _, tc := range tests {
		got := tc.side.String()
		if got != tc.expected {
			t.Errorf("PositionSide(%d).String() = %q, expected %q", int(tc.side), got, tc.expected)
		}
	}
}

func TestClientOrderIdString(t *testing.T) {
	id := ClientOrderId{Value: "O-001"}
	if id.String() != "O-001" {
		t.Errorf("ClientOrderId.String() = %q, expected %q", id.String(), "O-001")
	}
}

func TestStrategyIdString(t *testing.T) {
	id := StrategyId{Value: "EMACross-001"}
	if id.String() != "EMACross-001" {
		t.Errorf("StrategyId.String() = %q, expected %q", id.String(), "EMACross-001")
	}
}

func TestMarketOrderString(t *testing.T) {
	order := MarketOrder{
		ClientOrderId: ClientOrderId{Value: "O-001"},
		InstrumentId: NewInstrumentId(
			Symbol{Value: "6EH4"},
			Venue{Name: "XCME"},
		),
		Side:        OrderSideBuy,
		Quantity:    NewQuantity(10, 0),
		TimeInForce: TimeInForceGTC,
		Status:      OrderStatusInitialized,
		TsInit:      1000,
	}
	got := order.String()
	expected := "MarketOrder(O-001 6EH4.XCME BUY qty=10 GTC INITIALIZED)"
	if got != expected {
		t.Errorf("MarketOrder.String() = %q, expected %q", got, expected)
	}
}

func TestOrderFilledString(t *testing.T) {
	fill := OrderFilled{
		ClientOrderId: ClientOrderId{Value: "O-001"},
		InstrumentId: NewInstrumentId(
			Symbol{Value: "6EH4"},
			Venue{Name: "XCME"},
		),
		Side:      OrderSideBuy,
		Quantity:  NewQuantity(10, 0),
		FillPrice: NewPrice(1.12345, 5),
		TsEvent:   2000,
		TsInit:    2000,
	}
	got := fill.String()
	expected := "OrderFilled(O-001 6EH4.XCME BUY qty=10 @ 1.12345)"
	if got != expected {
		t.Errorf("OrderFilled.String() = %q, expected %q", got, expected)
	}
}
