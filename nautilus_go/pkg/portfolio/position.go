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

// Position tracks the net position for an instrument (netting OMS).
type Position struct {
	InstrumentId model.InstrumentId
	Side         model.PositionSide
	Quantity     model.Quantity
	SignedQty    float64
}

// NewPosition creates a new flat position for the given instrument.
func NewPosition(instrumentId model.InstrumentId) *Position {
	return &Position{
		InstrumentId: instrumentId,
		Side:         model.PositionSideFlat,
		Quantity:     model.NewQuantity(0, 0),
	}
}

// ApplyFill updates the position from a fill event.
func (p *Position) ApplyFill(fill model.OrderFilled) {
	fillQty := fill.Quantity.AsFloat64()
	switch fill.Side {
	case model.OrderSideBuy:
		p.SignedQty += fillQty
	case model.OrderSideSell:
		p.SignedQty -= fillQty
	}
	p.updateSide()
}

// IsFlat returns true if the position is flat (no exposure).
func (p *Position) IsFlat() bool {
	return p.Side == model.PositionSideFlat
}

// IsLong returns true if the position is net long.
func (p *Position) IsLong() bool {
	return p.Side == model.PositionSideLong
}

// IsShort returns true if the position is net short.
func (p *Position) IsShort() bool {
	return p.Side == model.PositionSideShort
}

func (p *Position) updateSide() {
	switch {
	case p.SignedQty > 0:
		p.Side = model.PositionSideLong
		p.Quantity = model.NewQuantity(p.SignedQty, 0)
	case p.SignedQty < 0:
		p.Side = model.PositionSideShort
		p.Quantity = model.NewQuantity(-p.SignedQty, 0)
	default:
		p.Side = model.PositionSideFlat
		p.Quantity = model.NewQuantity(0, 0)
	}
}
