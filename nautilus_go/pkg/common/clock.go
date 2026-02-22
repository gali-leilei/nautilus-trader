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

package common

import "github.com/nautechsystems/nautilus_go/pkg/model"

// Clock provides access to the current simulated or real time.
type Clock interface {
	TimestampNs() model.UnixNanos
	SetTime(ns model.UnixNanos)
}

// TestClock is a manually-controlled clock for backtesting.
type TestClock struct {
	currentNs model.UnixNanos
}

// NewTestClock creates a new TestClock set to zero.
func NewTestClock() *TestClock {
	return &TestClock{}
}

func (c *TestClock) TimestampNs() model.UnixNanos {
	return c.currentNs
}

func (c *TestClock) SetTime(ns model.UnixNanos) {
	c.currentNs = ns
}

// AdvanceTime moves the clock forward to the given timestamp.
func (c *TestClock) AdvanceTime(toNs model.UnixNanos) {
	if toNs > c.currentNs {
		c.currentNs = toNs
	}
}
