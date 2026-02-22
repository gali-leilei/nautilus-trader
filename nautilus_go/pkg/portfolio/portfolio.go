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

package portfolio

import "github.com/nautechsystems/nautilus_go/pkg/model"

// Portfolio tracks positions across multiple instruments.
type Portfolio struct {
	positions map[string]*Position
}

// NewPortfolio creates a new empty portfolio.
func NewPortfolio() *Portfolio {
	return &Portfolio{
		positions: make(map[string]*Position),
	}
}

// ApplyFill creates or updates the position for the filled instrument.
func (p *Portfolio) ApplyFill(fill model.OrderFilled) {
	key := fill.InstrumentId.String()
	pos, ok := p.positions[key]
	if !ok {
		pos = NewPosition(fill.InstrumentId)
		p.positions[key] = pos
	}
	pos.ApplyFill(fill)
}

// Position returns the position for an instrument, or nil if none.
func (p *Portfolio) Position(id model.InstrumentId) *Position {
	return p.positions[id.String()]
}

// IsFlat returns true if the position for the instrument is flat or absent.
func (p *Portfolio) IsFlat(id model.InstrumentId) bool {
	pos := p.positions[id.String()]
	return pos == nil || pos.IsFlat()
}

// IsNetLong returns true if the position for the instrument is net long.
func (p *Portfolio) IsNetLong(id model.InstrumentId) bool {
	pos := p.positions[id.String()]
	return pos != nil && pos.IsLong()
}

// IsNetShort returns true if the position for the instrument is net short.
func (p *Portfolio) IsNetShort(id model.InstrumentId) bool {
	pos := p.positions[id.String()]
	return pos != nil && pos.IsShort()
}
