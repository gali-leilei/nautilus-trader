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

// OrderSide represents the side of an order.
type OrderSide int

const (
	OrderSideBuy  OrderSide = 1
	OrderSideSell OrderSide = 2
)

func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(s))
	}
}

// TimeInForce represents the time-in-force for an order.
type TimeInForce int

const (
	TimeInForceGTC TimeInForce = 1
	TimeInForceIOC TimeInForce = 2
	TimeInForceFOK TimeInForce = 3
	TimeInForceDAY TimeInForce = 4
)

func (t TimeInForce) String() string {
	switch t {
	case TimeInForceGTC:
		return "GTC"
	case TimeInForceIOC:
		return "IOC"
	case TimeInForceFOK:
		return "FOK"
	case TimeInForceDAY:
		return "DAY"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(t))
	}
}

// OrderStatus represents the status of an order.
type OrderStatus int

const (
	OrderStatusInitialized OrderStatus = 1
	OrderStatusSubmitted   OrderStatus = 2
	OrderStatusAccepted    OrderStatus = 3
	OrderStatusFilled      OrderStatus = 4
	OrderStatusCanceled    OrderStatus = 5
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusInitialized:
		return "INITIALIZED"
	case OrderStatusSubmitted:
		return "SUBMITTED"
	case OrderStatusAccepted:
		return "ACCEPTED"
	case OrderStatusFilled:
		return "FILLED"
	case OrderStatusCanceled:
		return "CANCELED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(s))
	}
}

// PositionSide represents the side of a position.
type PositionSide int

const (
	PositionSideFlat  PositionSide = 1
	PositionSideLong  PositionSide = 2
	PositionSideShort PositionSide = 3
)

func (p PositionSide) String() string {
	switch p {
	case PositionSideFlat:
		return "FLAT"
	case PositionSideLong:
		return "LONG"
	case PositionSideShort:
		return "SHORT"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(p))
	}
}

// ClientOrderId uniquely identifies a client order.
type ClientOrderId struct {
	Value string
}

func (id ClientOrderId) String() string {
	return id.Value
}

// StrategyId identifies a strategy instance.
type StrategyId struct {
	Value string
}

func (id StrategyId) String() string {
	return id.Value
}

// MarketOrder represents a market order to be submitted.
type MarketOrder struct {
	ClientOrderId ClientOrderId
	InstrumentId  InstrumentId
	Side          OrderSide
	Quantity      Quantity
	TimeInForce   TimeInForce
	Status        OrderStatus
	TsInit        UnixNanos
}

func (o MarketOrder) String() string {
	return fmt.Sprintf(
		"MarketOrder(%s %s %s qty=%s %s %s)",
		o.ClientOrderId,
		o.InstrumentId,
		o.Side,
		o.Quantity,
		o.TimeInForce,
		o.Status,
	)
}

// OrderFilled represents a fill event for an order.
type OrderFilled struct {
	ClientOrderId ClientOrderId
	InstrumentId  InstrumentId
	Side          OrderSide
	Quantity      Quantity
	FillPrice     Price
	TsEvent       UnixNanos
	TsInit        UnixNanos
}

func (f OrderFilled) String() string {
	return fmt.Sprintf(
		"OrderFilled(%s %s %s qty=%s @ %s)",
		f.ClientOrderId,
		f.InstrumentId,
		f.Side,
		f.Quantity,
		f.FillPrice,
	)
}
