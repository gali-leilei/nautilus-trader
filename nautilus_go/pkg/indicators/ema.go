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
	"fmt"

	"github.com/nautechsystems/nautilus_go/pkg/model"
)

// ExponentialMovingAverage calculates an EMA over a price series.
type ExponentialMovingAverage struct {
	Period int
	Alpha  float64
	Value  float64
	Count  int
}

// NewExponentialMovingAverage creates a new EMA with the given period.
// Panics if period < 1.
func NewExponentialMovingAverage(period int) *ExponentialMovingAverage {
	if period < 1 {
		panic(fmt.Sprintf(
			"invalid EMA period: %d, must be >= 1", period,
		))
	}
	return &ExponentialMovingAverage{
		Period: period,
		Alpha:  2.0 / float64(period+1),
	}
}

// Name returns the indicator name with period.
func (ema *ExponentialMovingAverage) Name() string {
	return fmt.Sprintf("EMA(%d)", ema.Period)
}

// HasInputs returns true if the EMA has received any input.
func (ema *ExponentialMovingAverage) HasInputs() bool {
	return ema.Count > 0
}

// Initialized returns true when enough data has been received.
func (ema *ExponentialMovingAverage) Initialized() bool {
	return ema.Count >= ema.Period
}

// UpdateRaw updates the EMA with a raw float64 value.
func (ema *ExponentialMovingAverage) UpdateRaw(value float64) {
	if ema.Count == 0 {
		ema.Value = value
	} else {
		ema.Value = ema.Alpha*value + (1.0-ema.Alpha)*ema.Value
	}
	ema.Count++
}

// HandleBar updates the EMA with the bar's close price.
func (ema *ExponentialMovingAverage) HandleBar(bar model.Bar) {
	ema.UpdateRaw(bar.Close.AsFloat64())
}

// Reset zeroes all state.
func (ema *ExponentialMovingAverage) Reset() {
	ema.Value = 0
	ema.Count = 0
}
