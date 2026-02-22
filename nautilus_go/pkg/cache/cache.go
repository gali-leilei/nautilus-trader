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

package cache

import "github.com/nautechsystems/nautilus_go/pkg/model"

// Cache provides in-memory storage for instruments and bars.
type Cache struct {
	instruments map[string]model.Instrument
	bars        map[string][]model.Bar
}

// NewCache creates a new empty Cache.
func NewCache() *Cache {
	return &Cache{
		instruments: make(map[string]model.Instrument),
		bars:        make(map[string][]model.Bar),
	}
}

// AddInstrument stores an instrument keyed by its ID.
func (c *Cache) AddInstrument(inst model.Instrument) {
	c.instruments[inst.ID().String()] = inst
}

// Instrument retrieves an instrument by ID, or nil if not found.
func (c *Cache) Instrument(id model.InstrumentId) model.Instrument {
	return c.instruments[id.String()]
}

// AddBar appends a bar to the cache, keyed by its BarType string.
func (c *Cache) AddBar(bar model.Bar) {
	key := bar.BarType.String()
	c.bars[key] = append(c.bars[key], bar)
}

// Bar returns the last bar for the given BarType, or nil if none.
func (c *Cache) Bar(bt model.BarType) *model.Bar {
	key := bt.String()
	bars := c.bars[key]
	if len(bars) == 0 {
		return nil
	}
	return &bars[len(bars)-1]
}

// Bars returns all bars for the given BarType.
func (c *Cache) Bars(bt model.BarType) []model.Bar {
	return c.bars[bt.String()]
}

// BarCount returns the number of bars cached for the given BarType.
func (c *Cache) BarCount(bt model.BarType) int {
	return len(c.bars[bt.String()])
}
