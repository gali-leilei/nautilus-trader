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

// BarHandler is a callback function invoked when a bar is published.
type BarHandler func(model.Bar)

// MessageBus provides topic-based pub/sub for bars.
type MessageBus struct {
	barSubs map[string][]BarHandler
}

// NewMessageBus creates a new MessageBus.
func NewMessageBus() *MessageBus {
	return &MessageBus{
		barSubs: make(map[string][]BarHandler),
	}
}

// SubscribeBars registers a handler for the given topic.
func (mb *MessageBus) SubscribeBars(topic string, handler BarHandler) {
	mb.barSubs[topic] = append(mb.barSubs[topic], handler)
}

// PublishBar publishes a bar to all handlers subscribed to the topic.
func (mb *MessageBus) PublishBar(topic string, bar model.Bar) {
	for _, handler := range mb.barSubs[topic] {
		handler(bar)
	}
}
