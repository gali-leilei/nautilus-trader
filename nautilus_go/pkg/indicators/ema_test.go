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

package indicators

import (
	"math"
	"testing"

	"github.com/nautechsystems/nautilus_go/pkg/model"
)

// Interface compliance check
var _ Indicator = (*ExponentialMovingAverage)(nil)

func TestEMAAlpha(t *testing.T) {
	ema := NewExponentialMovingAverage(10)
	expected := 2.0 / 11.0
	if math.Abs(ema.Alpha-expected) > 1e-15 {
		t.Errorf("Alpha = %v, expected %v", ema.Alpha, expected)
	}
}

func TestEMAPeriod1(t *testing.T) {
	ema := NewExponentialMovingAverage(1)
	if ema.Alpha != 1.0 {
		t.Errorf("Period=1 Alpha = %v, expected 1.0", ema.Alpha)
	}

	// Period=1 should track the latest sample exactly
	values := []float64{1.0, 5.0, 3.0, 10.0}
	for _, v := range values {
		ema.UpdateRaw(v)
		if ema.Value != v {
			t.Errorf("Period=1 after input %v: Value = %v, expected %v", v, ema.Value, v)
		}
	}
}

func TestEMAFeed10Values(t *testing.T) {
	ema := NewExponentialMovingAverage(10)
	for i := 1; i <= 10; i++ {
		ema.UpdateRaw(float64(i))
	}

	// Expected value matches the Rust implementation
	expected := 6.239368480121215
	if math.Abs(ema.Value-expected) > 1e-12 {
		t.Errorf("EMA(10) after 1..10: Value = %.15f, expected %.15f", ema.Value, expected)
	}
}

func TestEMAInitialization(t *testing.T) {
	ema := NewExponentialMovingAverage(10)

	if ema.HasInputs() {
		t.Error("HasInputs() should be false before any input")
	}
	if ema.Initialized() {
		t.Error("Initialized() should be false before period inputs")
	}

	for i := 1; i <= 9; i++ {
		ema.UpdateRaw(float64(i))
		if !ema.HasInputs() {
			t.Errorf("HasInputs() should be true after %d inputs", i)
		}
		if ema.Initialized() {
			t.Errorf("Initialized() should be false after %d inputs (< period 10)", i)
		}
	}

	ema.UpdateRaw(10.0)
	if !ema.Initialized() {
		t.Error("Initialized() should be true after 10 inputs (= period)")
	}
}

func TestEMAName(t *testing.T) {
	ema := NewExponentialMovingAverage(20)
	if ema.Name() != "EMA(20)" {
		t.Errorf("Name() = %q, expected %q", ema.Name(), "EMA(20)")
	}
}

func makeTestBarType() model.BarType {
	bt, _ := model.ParseBarType("6EH4.XCME-1-MINUTE-LAST-EXTERNAL")
	return bt
}

func TestEMAHandleBar(t *testing.T) {
	ema := NewExponentialMovingAverage(3)
	bt := makeTestBarType()

	bar := model.Bar{
		BarType: bt,
		Open:    model.NewPrice(1.0, 5),
		High:    model.NewPrice(2.0, 5),
		Low:     model.NewPrice(0.5, 5),
		Close:   model.NewPrice(1.5, 5),
		Volume:  model.NewQuantity(100, 0),
	}
	ema.HandleBar(bar)

	if ema.Count != 1 {
		t.Errorf("Count = %d, expected 1", ema.Count)
	}
	// First bar sets value directly to close price
	if math.Abs(ema.Value-1.5) > 1e-10 {
		t.Errorf("Value = %v, expected 1.5", ema.Value)
	}
}

func TestEMAReset(t *testing.T) {
	ema := NewExponentialMovingAverage(10)
	for i := 1; i <= 10; i++ {
		ema.UpdateRaw(float64(i))
	}

	ema.Reset()
	if ema.Value != 0 {
		t.Errorf("after Reset() Value = %v, expected 0", ema.Value)
	}
	if ema.Count != 0 {
		t.Errorf("after Reset() Count = %d, expected 0", ema.Count)
	}
	if ema.HasInputs() {
		t.Error("after Reset() HasInputs() should be false")
	}
	if ema.Initialized() {
		t.Error("after Reset() Initialized() should be false")
	}

	// Can be reused after reset
	ema.UpdateRaw(5.0)
	if ema.Value != 5.0 {
		t.Errorf("after Reset()+UpdateRaw(5): Value = %v, expected 5", ema.Value)
	}
}

func TestEMAPanicOnZeroPeriod(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for period=0")
		}
	}()
	NewExponentialMovingAverage(0)
}
