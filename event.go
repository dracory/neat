package neat

import (
	"sync"
)

// EventBus is a lightweight internal event bus for model lifecycle events.
type EventBus struct {
	mu        sync.RWMutex
	listeners map[string][]EventHandler
}

// EventHandler is a function that handles an event.
type EventHandler func(event any)

// NewEventBus creates a new EventBus.
// It initializes an empty event bus with no registered listeners.
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventHandler),
	}
}

// Listen registers a handler for the given event name.
// The handler will be called whenever the event is dispatched.
func (e *EventBus) Listen(eventName string, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners[eventName] = append(e.listeners[eventName], handler)
}

// Dispatch dispatches an event to all registered listeners.
// It calls each handler synchronously in the order they were registered.
func (e *EventBus) Dispatch(eventName string, event any) {
	e.mu.RLock()
	handlers := e.listeners[eventName]
	e.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// Forget removes all listeners for the given event name.
// This clears all handlers registered for the specified event.
func (e *EventBus) Forget(eventName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, eventName)
}

// Event names for model lifecycle events.
const (
	EventCreating  = "model.creating"
	EventCreated   = "model.created"
	EventUpdating  = "model.updating"
	EventUpdated   = "model.updated"
	EventSaving    = "model.saving"
	EventSaved     = "model.saved"
	EventDeleting  = "model.deleting"
	EventDeleted   = "model.deleted"
	EventRestoring = "model.restoring"
	EventRestored  = "model.restored"
)
