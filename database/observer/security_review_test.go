package observer

// Security-review proof test for docs/security-review.md Finding #9.
// FAILS on the current implementation: a panicking observer callback
// propagates the panic out of Dispatch instead of being converted to an
// error, crashing the host process mid-transaction.

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
)

type panicObserver struct{}

func (panicObserver) Created(orm.Event) error      { panic("boom in Created") }
func (panicObserver) Updated(orm.Event) error      { return nil }
func (panicObserver) Deleted(orm.Event) error      { return nil }
func (panicObserver) ForceDeleted(orm.Event) error { return nil }

type testModel struct{ ID int }

// Finding #9: Dispatch must not let an observer panic escape to the caller.
func TestSecurityReviewObserverPanicContained(t *testing.T) {
	d := NewDispatcher(nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("observer panic propagated to caller instead of returning an error: %v", r)
		}
	}()

	// Contract: the observer panic is contained — Dispatch logs it and
	// continues to other observers, returning nil. If the panic escaped,
	// the deferred recover above fails the test (and without it the panic
	// would crash the whole test binary).
	_ = d.Dispatch(
		context.Background(),
		testModel{ID: 1},
		[]orm.ModelToObserver{{Model: testModel{}, Observer: panicObserver{}}},
		nil, nil, nil, nil,
		orm.EventCreated,
	)
}
