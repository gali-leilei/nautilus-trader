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
	"github.com/nautechsystems/nautilus_go/pkg/model"
)

// Strategy is the interface that user strategies must implement.
type Strategy interface {
	OnStart(ctx *StrategyContext)
	OnBar(ctx *StrategyContext, bar model.Bar)
	OnStop(ctx *StrategyContext)
}

// StrategyContext provides services to a strategy during execution.
type StrategyContext struct {
	Clock  common.Clock
	Cache  *cache.Cache
	MsgBus *common.MessageBus
	Logger *slog.Logger
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
