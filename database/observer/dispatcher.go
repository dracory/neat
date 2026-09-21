package observer

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dracory/neat/contracts/database/orm"
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
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
	eventType orm.EventType,
) error {
	if len(observers) == 0 {
		return nil
	}

	// Create event
	event := NewEvent(ctx, model, original, attributes, dirty, query, eventType)

	// Find matching observers for this model
	var matchingObservers []orm.Observer
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}

	for _, mt := range observers {
		observerType := reflect.TypeOf(mt.Model)
		if observerType.Kind() == reflect.Pointer {
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
func (d *Dispatcher) dispatchToObserver(event *Event, observer orm.Observer, eventType orm.EventType) error {
	switch eventType {
	case orm.EventCreating:
		if obs, ok := observer.(orm.ObserverWithCreating); ok {
			return obs.Creating(event)
		}
	case orm.EventCreated:
		return observer.Created(event)
	case orm.EventUpdating:
		if obs, ok := observer.(orm.ObserverWithUpdating); ok {
			return obs.Updating(event)
		}
	case orm.EventUpdated:
		return observer.Updated(event)
	case orm.EventSaving:
		if obs, ok := observer.(orm.ObserverWithSaving); ok {
			return obs.Saving(event)
		}
	case orm.EventSaved:
		if obs, ok := observer.(orm.ObserverWithSaved); ok {
			return obs.Saved(event)
		}
	case orm.EventDeleting:
		if obs, ok := observer.(orm.ObserverWithDeleting); ok {
			return obs.Deleting(event)
		}
	case orm.EventDeleted:
		return observer.Deleted(event)
	case orm.EventForceDeleting:
		if obs, ok := observer.(orm.ObserverWithForceDeleting); ok {
			return obs.ForceDeleting(event)
		}
	case orm.EventForceDeleted:
		return observer.ForceDeleted(event)
	case orm.EventSoftDeleteRestoring:
		if obs, ok := observer.(orm.ObserverWithRestoring); ok {
			return obs.Restoring(event)
		}
	case orm.EventSoftDeleteRestored:
		if obs, ok := observer.(orm.ObserverWithRestored); ok {
			return obs.Restored(event)
		}
	case orm.EventRetrieved:
		if obs, ok := observer.(orm.ObserverWithRetrieved); ok {
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
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventCreating)
}

// DispatchCreated dispatches the created event.
func (d *Dispatcher) DispatchCreated(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventCreated)
}

// DispatchUpdating dispatches the updating event.
func (d *Dispatcher) DispatchUpdating(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventUpdating)
}

// DispatchUpdated dispatches the updated event.
func (d *Dispatcher) DispatchUpdated(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventUpdated)
}

// DispatchSaving dispatches the saving event.
func (d *Dispatcher) DispatchSaving(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventSaving)
}

// DispatchSaved dispatches the saved event.
func (d *Dispatcher) DispatchSaved(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventSaved)
}

// DispatchDeleting dispatches the deleting event.
func (d *Dispatcher) DispatchDeleting(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventDeleting)
}

// DispatchDeleted dispatches the deleted event.
func (d *Dispatcher) DispatchDeleted(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventDeleted)
}

// DispatchForceDeleting dispatches the force deleting event.
func (d *Dispatcher) DispatchForceDeleting(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventForceDeleting)
}

// DispatchForceDeleted dispatches the force deleted event.
func (d *Dispatcher) DispatchForceDeleted(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventForceDeleted)
}

// DispatchRestoring dispatches the restoring event.
func (d *Dispatcher) DispatchRestoring(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventSoftDeleteRestoring)
}

// DispatchRestored dispatches the restored event.
func (d *Dispatcher) DispatchRestored(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventSoftDeleteRestored)
}

// DispatchRetrieved dispatches the retrieved event.
func (d *Dispatcher) DispatchRetrieved(
	ctx context.Context,
	model any,
	observers []orm.ModelToObserver,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query orm.Query,
) error {
	return d.Dispatch(ctx, model, observers, original, attributes, dirty, query, orm.EventRetrieved)
}
