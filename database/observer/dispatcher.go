package observer

import (
	"context"
	"fmt"
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
)

// Dispatcher handles event dispatching for model lifecycle events.
type Dispatcher struct {
	log log.Log
}

// NewDispatcher creates a new Dispatcher instance.
func NewDispatcher(log log.Log) *Dispatcher {
	return &Dispatcher{
		log: log,
	}
}

// Dispatch dispatches an event to all registered observers for the given model.
func (d *Dispatcher) Dispatch(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
	eventType contractsorm.EventType,
) error {
	if len(observers) == 0 {
		return nil
	}

	// Create event
	event := NewEvent(ctx, model, original, attributes, dirty, query, eventType)

	// Find matching observers for this model
	var matchingObservers []contractsorm.Observer
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for _, mt := range observers {
		observerType := reflect.TypeOf(mt.Model)
		if observerType.Kind() == reflect.Ptr {
			observerType = observerType.Elem()
		}

		// Check if observer matches the model type
		if observerType == modelType {
			matchingObservers = append(matchingObservers, mt.Observer)
		}
	}

	// Dispatch event to matching observers
	for _, observer := range matchingObservers {
		if err := d.dispatchToObserver(event, observer, eventType); err != nil {
			d.log.Errorf("[Dispatcher] Error dispatching event %s: %v", eventType, err)
			// Continue with other observers even if one fails
		}
	}

	return nil
}

// dispatchToObserver dispatches an event to a specific observer based on the event type.
func (d *Dispatcher) dispatchToObserver(event *Event, observer contractsorm.Observer, eventType contractsorm.EventType) error {
	switch eventType {
	case contractsorm.EventCreating:
		if obs, ok := observer.(contractsorm.ObserverWithCreating); ok {
			return obs.Creating(event)
		}
	case contractsorm.EventCreated:
		return observer.Created(event)
	case contractsorm.EventUpdating:
		if obs, ok := observer.(contractsorm.ObserverWithUpdating); ok {
			return obs.Updating(event)
		}
	case contractsorm.EventUpdated:
		return observer.Updated(event)
	case contractsorm.EventSaving:
		if obs, ok := observer.(contractsorm.ObserverWithSaving); ok {
			return obs.Saving(event)
		}
	case contractsorm.EventSaved:
		if obs, ok := observer.(contractsorm.ObserverWithSaved); ok {
			return obs.Saved(event)
		}
	case contractsorm.EventDeleting:
		if obs, ok := observer.(contractsorm.ObserverWithDeleting); ok {
			return obs.Deleting(event)
		}
	case contractsorm.EventDeleted:
		return observer.Deleted(event)
	case contractsorm.EventForceDeleting:
		if obs, ok := observer.(contractsorm.ObserverWithForceDeleting); ok {
			return obs.ForceDeleting(event)
		}
	case contractsorm.EventForceDeleted:
		return observer.ForceDeleted(event)
	case contractsorm.EventRestoring:
		if obs, ok := observer.(contractsorm.ObserverWithRestoring); ok {
			return obs.Restoring(event)
		}
	case contractsorm.EventRestored:
		if obs, ok := observer.(contractsorm.ObserverWithRestored); ok {
			return obs.Restored(event)
		}
	case contractsorm.EventRetrieved:
		if obs, ok := observer.(contractsorm.ObserverWithRetrieved); ok {
			return obs.Retrieved(event)
		}
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}

	return nil
}

// DispatchCreating dispatches the creating event.
func (d *Dispatcher) DispatchCreating(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventCreating)
}

// DispatchCreated dispatches the created event.
func (d *Dispatcher) DispatchCreated(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventCreated)
}

// DispatchUpdating dispatches the updating event.
func (d *Dispatcher) DispatchUpdating(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventUpdating)
}

// DispatchUpdated dispatches the updated event.
func (d *Dispatcher) DispatchUpdated(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventUpdated)
}

// DispatchSaving dispatches the saving event.
func (d *Dispatcher) DispatchSaving(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventSaving)
}

// DispatchSaved dispatches the saved event.
func (d *Dispatcher) DispatchSaved(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventSaved)
}

// DispatchDeleting dispatches the deleting event.
func (d *Dispatcher) DispatchDeleting(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventDeleting)
}

// DispatchDeleted dispatches the deleted event.
func (d *Dispatcher) DispatchDeleted(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventDeleted)
}

// DispatchForceDeleting dispatches the force deleting event.
func (d *Dispatcher) DispatchForceDeleting(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventForceDeleting)
}

// DispatchForceDeleted dispatches the force deleted event.
func (d *Dispatcher) DispatchForceDeleted(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventForceDeleted)
}

// DispatchRestoring dispatches the restoring event.
func (d *Dispatcher) DispatchRestoring(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventRestoring)
}

// DispatchRestored dispatches the restored event.
func (d *Dispatcher) DispatchRestored(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventRestored)
}

// DispatchRetrieved dispatches the retrieved event.
func (d *Dispatcher) DispatchRetrieved(
	ctx context.Context,
	model any,
	observers []contractsorm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, contractsorm.EventRetrieved)
}
