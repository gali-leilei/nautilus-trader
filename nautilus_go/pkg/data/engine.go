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

package data

import (
	"log/slog"

	"github.com/nautechsystems/nautilus_go/pkg/cache"
	"github.com/nautechsystems/nautilus_go/pkg/common"
	"github.com/nautechsystems/nautilus_go/pkg/model"
)

// DataEngine validates, caches, and publishes data.
type DataEngine struct {
	cache            *cache.Cache
	msgbus           *common.MessageBus
	logger           *slog.Logger
	validateSequence bool
}

// NewDataEngine creates a new DataEngine.
func NewDataEngine(
	c *cache.Cache,
	msgbus *common.MessageBus,
	logger *slog.Logger,
) *DataEngine {
	return &DataEngine{
		cache:            c,
		msgbus:           msgbus,
		logger:           logger,
		validateSequence: true,
	}
}

// ProcessBar validates a bar, adds it to the cache, and publishes it.
func (de *DataEngine) ProcessBar(bar model.Bar) {
	if de.validateSequence {
		lastBar := de.cache.Bar(bar.BarType)
		if lastBar != nil {
			if bar.TsEvent < lastBar.TsEvent {
				de.logger.Warn(
					"Bar was prior to last bar ts_event",
					"bar", bar.String(),
					"last_ts_event", lastBar.TsEvent,
				)
				return
			}
			if bar.TsInit < lastBar.TsInit {
				de.logger.Warn(
					"Bar was prior to last bar ts_init",
					"bar", bar.String(),
					"last_ts_init", lastBar.TsInit,
				)
				return
			}
		}
	}

	de.cache.AddBar(bar)

	topic := bar.BarType.Topic()
	de.msgbus.PublishBar(topic, bar)
}
