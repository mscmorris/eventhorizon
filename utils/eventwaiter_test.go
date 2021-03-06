// Copyright (c) 2017 - Max Ekman <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"reflect"
	"testing"

	"time"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/mocks"
)

func TestEventWaiter(t *testing.T) {
	w := NewEventWaiter()

	// Event should match when waiting.
	expectedEvent := eh.NewEventForAggregate(mocks.EventType, nil, mocks.AggregateType, eh.NewUUID(), 1)
	go func() {
		time.Sleep(time.Millisecond)
		if err := w.Notify(context.Background(), expectedEvent); err != nil {
			t.Error(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	event, err := w.Wait(context.Background(), func(event eh.Event) bool {
		if event.EventType() == mocks.EventType {
			return true
		}
		return false
	})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(event, expectedEvent) {
		t.Error("the event should be correct:", event)
	}
	if len(w.waits) > 0 {
		t.Error("there should be no open waits")
	}

	// Other events should not match.
	otherEvent := eh.NewEventForAggregate(mocks.EventOtherType, nil, mocks.AggregateType, eh.NewUUID(), 1)
	go func() {
		time.Sleep(time.Millisecond)
		if err := w.Notify(context.Background(), otherEvent); err != nil {
			t.Error(err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	event, err = w.Wait(ctx, func(event eh.Event) bool {
		if event.EventType() == mocks.EventType {
			return true
		}
		return false
	})
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Error("there should be a context deadline exceeded error")
	}
	if event != nil {
		t.Error("the event should be nil:", event)
	}
	if len(w.waits) > 0 {
		t.Error("there should be no open waits")
	}
}
