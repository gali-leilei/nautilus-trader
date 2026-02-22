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

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nautechsystems/nautilus_go/pkg/backtest"
	"github.com/nautechsystems/nautilus_go/pkg/model"
	"github.com/nautechsystems/nautilus_go/pkg/trading"
)

// loadBarsFromCSV reads semicolon-delimited OHLCV bars from a CSV file.
func loadBarsFromCSV(
	path string,
	barType model.BarType,
	pricePrecision uint8,
	sizePrecision uint8,
) ([]model.Bar, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, fmt.Errorf("open CSV: %w", e)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'

	records, e := reader.ReadAll()
	if e != nil {
		return nil, fmt.Errorf("read CSV: %w", e)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	bars := make([]model.Bar, 0, len(records)-1)
	for i, row := range records[1:] { // skip header
		if len(row) < 6 {
			return nil, fmt.Errorf("row %d: expected 6+ columns, found %d", i+1, len(row))
		}

		ts, e := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(row[0]))
		if e != nil {
			return nil, fmt.Errorf("row %d: parse timestamp: %w", i+1, e)
		}
		tsNanos := model.UnixNanos(ts.UnixNano())

		open, e := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if e != nil {
			return nil, fmt.Errorf("row %d: parse open: %w", i+1, e)
		}
		high, e := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		if e != nil {
			return nil, fmt.Errorf("row %d: parse high: %w", i+1, e)
		}
		low, e := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		if e != nil {
			return nil, fmt.Errorf("row %d: parse low: %w", i+1, e)
		}
		close_, e := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if e != nil {
			return nil, fmt.Errorf("row %d: parse close: %w", i+1, e)
		}
		vol, e := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
		if e != nil {
			return nil, fmt.Errorf("row %d: parse volume: %w", i+1, e)
		}

		bar := model.Bar{
			BarType: barType,
			Open:    model.NewPrice(open, pricePrecision),
			High:    model.NewPrice(high, pricePrecision),
			Low:     model.NewPrice(low, pricePrecision),
			Close:   model.NewPrice(close_, pricePrecision),
			Volume:  model.NewQuantity(vol, sizePrecision),
			TsEvent: tsNanos,
			TsInit:  tsNanos,
		}
		bars = append(bars, bar)
	}

	return bars, nil
}

// DemoStrategy counts all processed bars and logs progress.
type DemoStrategy struct {
	primaryBarType model.BarType
	barsProcessed  int
	startTime      time.Time
}

func (s *DemoStrategy) OnStart(ctx *trading.StrategyContext) {
	s.startTime = time.Now()
	ctx.Logger.Info(fmt.Sprintf("Strategy started at: %s", s.startTime.Format(time.RFC3339)))
	ctx.SubscribeBars(s.primaryBarType, s)
}

func (s *DemoStrategy) OnBar(ctx *trading.StrategyContext, bar model.Bar) {
	s.barsProcessed++
	ctx.Logger.Info(
		fmt.Sprintf("Processed bars: %d", s.barsProcessed),
		"close", bar.Close.String(),
	)
}

func (s *DemoStrategy) OnStop(ctx *trading.StrategyContext) {
	endTime := time.Now()
	ctx.Logger.Info(fmt.Sprintf("Strategy finished at: %s", endTime.Format(time.RFC3339)))
	ctx.Logger.Info(fmt.Sprintf("Total bars processed: %d", s.barsProcessed))
}

func main() {
	// Step 1: Configure and create backtest engine
	engineConfig := backtest.EngineConfig{
		TraderID: "BACKTEST_TRADER-001",
		LogLevel: "DEBUG",
	}
	engine := backtest.NewBacktestEngine(engineConfig)

	// Step 2: Define exchange and add it to the engine
	xcme := model.Venue{Name: "XCME"}
	engine.AddVenue(backtest.VenueConfig{
		Venue:       xcme,
		OmsType:     model.OmsTypeNetting,
		AccountType: model.AccountTypeMargin,
		StartingBalances: []model.Money{
			model.NewMoney(1_000_000, model.USD),
		},
		BaseCurrency: model.USD,
	})

	// Step 3: Create instrument and add it to the engine
	instrument := &model.FuturesContract{
		InstrumentId: model.NewInstrumentId(
			model.Symbol{Value: "6EH4"},
			model.Venue{Name: "XCME"},
		),
		RawSymbol:      model.Symbol{Value: "6EH4"},
		AssetClass:     model.AssetClassFX,
		QuoteCurr:      model.USD,
		Underlying:     "EUR/USD",
		PricePrec:      5,
		SizePrec:       0,
		PriceIncrement: model.NewPrice(0.00005, 5),
		SizeIncrement:  model.NewQuantity(1, 0),
		Multiplier:     model.NewQuantity(125000, 0),
		LotSize:        model.NewQuantity(1, 0),
		Expiry:         time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
	}
	engine.AddInstrument(instrument)

	// Step 4: Load bar data from CSV
	barTypeStr := fmt.Sprintf("%s-1-MINUTE-LAST-EXTERNAL", instrument.ID())
	barType, e := model.ParseBarType(barTypeStr)
	if e != nil {
		log.Fatalf("Failed to parse BarType: %v", e)
	}

	bars, e := loadBarsFromCSV(
		"6EH4.XCME_1min_bars.csv",
		barType,
		instrument.PricePrecision(),
		instrument.SizePrecision(),
	)
	if e != nil {
		log.Fatalf("Failed to load bars: %v", e)
	}
	engine.AddData(bars)

	// Step 5: Create strategy and add it to the engine
	strategy := &DemoStrategy{primaryBarType: barType}
	engine.AddStrategy(strategy)

	// Step 6: Run the backtest
	engine.Run()

	// Step 7: Release resources
	engine.Dispose()
}
