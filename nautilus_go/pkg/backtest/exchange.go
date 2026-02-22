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

	"github.com/nautechsystems/nautilus_go/pkg/model"
	"github.com/nautechsystems/nautilus_go/pkg/portfolio"
)

// FillHandler is a callback invoked when an order is filled.
type FillHandler func(model.OrderFilled)

// SimulatedExchange simulates order matching for backtesting.
type SimulatedExchange struct {
	Venue        model.Venue
	Portfolio    *portfolio.Portfolio
	logger       *slog.Logger
	currentBar   *model.Bar
	pending      []model.MarketOrder
	fillHandlers []FillHandler
}

// NewSimulatedExchange creates a new SimulatedExchange.
func NewSimulatedExchange(
	venue model.Venue,
	port *portfolio.Portfolio,
	logger *slog.Logger,
) *SimulatedExchange {
	return &SimulatedExchange{
		Venue:     venue,
		Portfolio: port,
		logger:    logger,
	}
}

// OnFill registers a fill callback handler.
func (ex *SimulatedExchange) OnFill(handler FillHandler) {
	ex.fillHandlers = append(ex.fillHandlers, handler)
}

// ProcessBar updates the current market prices from the bar.
func (ex *SimulatedExchange) ProcessBar(bar model.Bar) {
	ex.currentBar = &bar
}

// SubmitOrder queues a market order for execution.
func (ex *SimulatedExchange) SubmitOrder(order model.MarketOrder) {
	order.Status = model.OrderStatusSubmitted
	ex.pending = append(ex.pending, order)
	ex.logger.Debug(
		"Order submitted",
		"client_order_id", order.ClientOrderId.Value,
		"instrument", order.InstrumentId,
		"side", order.Side,
		"qty", order.Quantity,
	)
}

// ProcessOrders fills all pending market orders at the current bar close.
func (ex *SimulatedExchange) ProcessOrders() {
	if ex.currentBar == nil {
		return
	}

	for _, order := range ex.pending {
		fill := model.OrderFilled{
			ClientOrderId: order.ClientOrderId,
			InstrumentId:  order.InstrumentId,
			Side:          order.Side,
			Quantity:      order.Quantity,
			FillPrice:     ex.currentBar.Close,
			TsEvent:       ex.currentBar.TsInit,
			TsInit:        ex.currentBar.TsInit,
		}

		ex.Portfolio.ApplyFill(fill)

		for _, handler := range ex.fillHandlers {
			handler(fill)
		}

		ex.logger.Debug(
			"Order filled",
			"client_order_id", fill.ClientOrderId.Value,
			"price", fill.FillPrice,
			"side", fill.Side,
			"qty", fill.Quantity,
		)
	}
	ex.pending = ex.pending[:0]
}

// CancelAllOrders removes all pending orders for the given instrument.
func (ex *SimulatedExchange) CancelAllOrders(instrumentId model.InstrumentId) {
	key := instrumentId.String()
	filtered := ex.pending[:0]
	for _, order := range ex.pending {
		if order.InstrumentId.String() != key {
			filtered = append(filtered, order)
		}
	}
	ex.pending = filtered
}
