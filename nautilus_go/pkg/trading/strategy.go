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
	"log/slog"

	"github.com/nautechsystems/nautilus_go/pkg/cache"
	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/indicators"
	"github.com/nautechsystems/nautilus_go/pkg/model"
	"github.com/nautechsystems/nautilus_go/pkg/portfolio"
)

// Exchange is the interface for order submission (avoids import cycle).
type Exchange interface {
	SubmitOrder(order model.MarketOrder)
	CancelAllOrders(instrumentId model.InstrumentId)
}

// Strategy is the interface that user strategies must implement.
type Strategy interface {
	OnStart(ctx *StrategyContext)
	OnBar(ctx *StrategyContext, bar model.Bar)
	OnStop(ctx *StrategyContext)
}

// indicatorBarReg pairs an indicator with the bar type it's registered for.
type indicatorBarReg struct {
	barType   model.BarType
	indicator indicators.Indicator
}

// StrategyContext provides services to a strategy during execution.
type StrategyContext struct {
	Clock        common.Clock
	Cache        *cache.Cache
	MsgBus       *common.MessageBus
	Logger       *slog.Logger
	Portfolio    *portfolio.Portfolio
	OrderFactory *OrderFactory
	exchange     Exchange
	indicators   []indicatorBarReg
}

// SetExchange sets the exchange for order routing (called by engine).
func (ctx *StrategyContext) SetExchange(exchange Exchange) {
	ctx.exchange = exchange
}

// SubscribeBars registers the strategy's OnBar handler for a BarType.
func (ctx *StrategyContext) SubscribeBars(
	bt model.BarType,
	strategy Strategy,
) {
	topic := bt.Topic()
	ctx.MsgBus.SubscribeBars(topic, func(bar model.Bar) {
		strategy.OnBar(ctx, bar)
	})
}

// RegisterIndicatorForBars registers an indicator to receive bars.
func (ctx *StrategyContext) RegisterIndicatorForBars(
	bt model.BarType,
	indicator indicators.Indicator,
) {
	ctx.indicators = append(ctx.indicators, indicatorBarReg{
		barType:   bt,
		indicator: indicator,
	})
}

// IndicatorsInitialized returns true if all registered indicators are
// initialized. Returns false if no indicators are registered.
func (ctx *StrategyContext) IndicatorsInitialized() bool {
	if len(ctx.indicators) == 0 {
		return false
	}
	for _, reg := range ctx.indicators {
		if !reg.indicator.Initialized() {
			return false
		}
	}
	return true
}

// UpdateIndicators feeds a bar to all indicators registered for its type.
func (ctx *StrategyContext) UpdateIndicators(bar model.Bar) {
	barTypeStr := bar.BarType.String()
	for _, reg := range ctx.indicators {
		if reg.barType.String() == barTypeStr {
			reg.indicator.HandleBar(bar)
		}
	}
}

// SubmitOrder forwards a market order to the exchange.
func (ctx *StrategyContext) SubmitOrder(order model.MarketOrder) {
	if ctx.exchange == nil {
		return
	}
	ctx.exchange.SubmitOrder(order)
}

// CloseAllPositions creates a reverse market order to flatten the position.
func (ctx *StrategyContext) CloseAllPositions(
	instrumentId model.InstrumentId,
) {
	if ctx.Portfolio == nil || ctx.OrderFactory == nil || ctx.exchange == nil {
		return
	}
	pos := ctx.Portfolio.Position(instrumentId)
	if pos == nil || pos.IsFlat() {
		return
	}

	var side model.OrderSide
	if pos.IsLong() {
		side = model.OrderSideSell
	} else {
		side = model.OrderSideBuy
	}

	order := ctx.OrderFactory.Market(
		instrumentId,
		side,
		pos.Quantity,
		model.TimeInForceGTC,
	)
	ctx.exchange.SubmitOrder(order)
}

// CancelAllOrders cancels all pending orders for an instrument.
func (ctx *StrategyContext) CancelAllOrders(
	instrumentId model.InstrumentId,
) {
	if ctx.exchange == nil {
		return
	}
	ctx.exchange.CancelAllOrders(instrumentId)
}
