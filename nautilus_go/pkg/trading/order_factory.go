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
	"fmt"

	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/model"
)

// OrderFactory generates unique market orders with auto-incrementing IDs.
type OrderFactory struct {
	TraderId   model.TraderId
	StrategyId model.StrategyId
	clock      common.Clock
	counter    int
}

// NewOrderFactory creates a new OrderFactory.
func NewOrderFactory(
	traderId model.TraderId,
	strategyId model.StrategyId,
	clock common.Clock,
) *OrderFactory {
	return &OrderFactory{
		TraderId:   traderId,
		StrategyId: strategyId,
		clock:      clock,
	}
}

// Market creates a new market order with a unique client order ID.
func (f *OrderFactory) Market(
	instrumentId model.InstrumentId,
	side model.OrderSide,
	quantity model.Quantity,
	timeInForce model.TimeInForce,
) model.MarketOrder {
	f.counter++
	clientOrderId := model.ClientOrderId{
		Value: fmt.Sprintf(
			"O-%s-%s-%d",
			f.TraderId.Value,
			f.StrategyId.Value,
			f.counter,
		),
	}
	return model.MarketOrder{
		ClientOrderId: clientOrderId,
		InstrumentId:  instrumentId,
		Side:          side,
		Quantity:      quantity,
		TimeInForce:   timeInForce,
		Status:        model.OrderStatusInitialized,
		TsInit:        f.clock.TimestampNs(),
	}
}

// Reset resets the counter to 0.
func (f *OrderFactory) Reset() {
	f.counter = 0
}
