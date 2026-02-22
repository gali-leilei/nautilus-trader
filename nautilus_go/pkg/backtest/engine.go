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
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/nautechsystems/nautilus_go/pkg/cache"
	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/data"
	"github.com/nautechsystems/nautilus_go/pkg/model"
	"github.com/nautechsystems/nautilus_go/pkg/portfolio"
	"github.com/nautechsystems/nautilus_go/pkg/trading"
)

// BacktestEngine orchestrates a backtest run.
type BacktestEngine struct {
	config     EngineConfig
	clock      *common.TestClock
	msgbus     *common.MessageBus
	cache      *cache.Cache
	dataEngine *data.DataEngine
	logger     *slog.Logger
	venues     []VenueConfig
	strategies []trading.Strategy
	bars       []model.Bar
	exchanges  map[string]*SimulatedExchange
	portfolio  *portfolio.Portfolio
}

// NewBacktestEngine creates a new BacktestEngine with the given config.
func NewBacktestEngine(config EngineConfig) *BacktestEngine {
	level := parseLogLevel(config.LogLevel)
	logger := common.NewLogger(level)
	clock := common.NewTestClock()
	msgbus := common.NewMessageBus()
	c := cache.NewCache()
	de := data.NewDataEngine(c, msgbus, logger)

	return &BacktestEngine{
		config:     config,
		clock:      clock,
		msgbus:     msgbus,
		cache:      c,
		dataEngine: de,
		logger:     logger,
		exchanges:  make(map[string]*SimulatedExchange),
		portfolio:  portfolio.NewPortfolio(),
	}
}

// AddVenue registers a simulated venue configuration.
func (e *BacktestEngine) AddVenue(config VenueConfig) {
	e.venues = append(e.venues, config)
	ex := NewSimulatedExchange(config.Venue, e.portfolio, e.logger)
	e.exchanges[config.Venue.Name] = ex
	e.logger.Info(
		"Added venue",
		"venue", config.Venue.Name,
		"oms_type", config.OmsType,
		"account_type", config.AccountType,
	)
}

// AddInstrument adds an instrument to the cache.
func (e *BacktestEngine) AddInstrument(inst model.Instrument) {
	e.cache.AddInstrument(inst)
	e.logger.Info("Added instrument", "id", inst.ID())
}

// AddData adds bar data to the engine for replay.
func (e *BacktestEngine) AddData(bars []model.Bar) {
	e.bars = append(e.bars, bars...)
	e.logger.Info("Added bar data", "count", len(bars))
}

// AddStrategy registers a strategy to receive data during the run.
func (e *BacktestEngine) AddStrategy(strategy trading.Strategy) {
	e.strategies = append(e.strategies, strategy)
	e.logger.Info("Added strategy")
}

// Run executes the backtest.
func (e *BacktestEngine) Run() {
	e.logger.Info(
		"Running backtest engine",
		"trader_id", e.config.TraderID,
		"bars", len(e.bars),
		"strategies", len(e.strategies),
	)

	// Sort bars by TsInit for chronological replay
	sort.Slice(e.bars, func(i, j int) bool {
		return e.bars[i].TsInit < e.bars[j].TsInit
	})

	// Build strategy contexts and call OnStart
	contexts := make([]*trading.StrategyContext, len(e.strategies))
	for i, strat := range e.strategies {
		ctx := &trading.StrategyContext{
			Clock:    e.clock,
			Cache:    e.cache,
			MsgBus:   e.msgbus,
			Logger:   e.logger,
			Portfolio: e.portfolio,
		}

		// Create OrderFactory for each strategy
		traderId := model.TraderId{Value: e.config.TraderID}
		strategyId := model.StrategyId{
			Value: fmt.Sprintf("Strategy-%d", i),
		}
		ctx.OrderFactory = trading.NewOrderFactory(
			traderId, strategyId, e.clock,
		)

		// Attach the first exchange (if any) for order routing
		for _, ex := range e.exchanges {
			ctx.SetExchange(ex)
			break
		}

		contexts[i] = ctx
		strat.OnStart(ctx)
	}

	// Main backtest loop
	for _, bar := range e.bars {
		e.clock.SetTime(bar.TsInit)

		// Find exchange for this bar's venue
		venue := bar.BarType.InstrumentId.Venue.Name
		if ex, ok := e.exchanges[venue]; ok {
			ex.ProcessBar(bar)
		}

		// Update registered indicators before OnBar
		for _, ctx := range contexts {
			ctx.UpdateIndicators(bar)
		}

		// Cache + publish (triggers OnBar via message bus)
		e.dataEngine.ProcessBar(bar)

		// Fill queued orders after OnBar
		if ex, ok := e.exchanges[venue]; ok {
			ex.ProcessOrders()
		}
	}

	// Call OnStop for each strategy
	for i, strat := range e.strategies {
		strat.OnStop(contexts[i])
	}

	e.logger.Info("Backtest complete")
}

// Dispose releases engine resources.
func (e *BacktestEngine) Dispose() {
	e.bars = nil
	e.strategies = nil
	e.logger.Info("Engine disposed")
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
