// Code generated by "go generate". DO NOT EDIT.

package watchable_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/datawire/dlib/dlog"
	"github.com/telepresenceio/telepresence/rpc/v2/manager"
	"github.com/telepresenceio/telepresence/v2/cmd/traffic/cmd/manager/internal/watchable"
)

func assertAgentMapSnapshotEqual(t *testing.T, expected, actual watchable.AgentMapSnapshot, msgAndArgs ...interface{}) bool {
	t.Helper()

	expectedBytes, err := json.MarshalIndent(expected, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	actualBytes, err := json.MarshalIndent(actual, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	if !assert.Equal(t, string(expectedBytes), string(actualBytes)) {
		return false
	}

	for k := range actual.State {
		if !assertDeepCopies(t, expected.State[k], actual.State[k], msgAndArgs...) {
			return false
		}
	}

	for i := range actual.Updates {
		if expected.Updates[i].Value == nil {
			continue
		}
		if !assertDeepCopies(t, expected.Updates[i].Value, actual.Updates[i].Value, msgAndArgs...) {
			return false
		}
	}

	return true
}

func TestAgentMap_Close(t *testing.T) {
	// TODO
}

func TestAgentMap_Delete(t *testing.T) {
	var m watchable.AgentMap

	// Check that a delete on a zero map works
	m.Delete("a")
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{State: map[string]*manager.AgentInfo{}},
		watchable.AgentMapSnapshot{State: m.LoadAll()})

	// Check that a normal delete works
	m.Store("a", &manager.AgentInfo{Name: "a"})
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"a": &manager.AgentInfo{Name: "a"},
			},
		},
		watchable.AgentMapSnapshot{State: m.LoadAll()})
	m.Delete("a")
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{},
		},
		watchable.AgentMapSnapshot{State: m.LoadAll()})

	// Check that a repeated delete works
	m.Delete("a")
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{},
		},
		watchable.AgentMapSnapshot{State: m.LoadAll()})
}

func TestAgentMap_Load(t *testing.T) {
	var m watchable.AgentMap

	a := &manager.AgentInfo{Name: "value"}
	m.Store("k", a)

	// Check that a load returns a copy of the input object
	b, ok := m.Load("k")
	assert.True(t, ok)
	assertDeepCopies(t, a, b)
	m.Delete("k")

	// Check that a load returns nil after a delete
	c, ok := m.Load("k")
	assert.False(t, ok)
	assert.Nil(t, c)

	// Check that two sequential loads return distinct copies
	m.Store("k", a)
	d, ok := m.Load("k")
	assert.True(t, ok)
	e, ok := m.Load("k")
	assert.True(t, ok)
	assertDeepCopies(t, a, d)
	assertDeepCopies(t, a, e)
	assertDeepCopies(t, d, e)
}

func TestAgentMap_LoadAll(t *testing.T) {
	// TODO
}

func TestAgentMap_LoadAndDelete(t *testing.T) {
	var m watchable.AgentMap

	a := &manager.AgentInfo{Name: "value"}
	m.Store("k", a)

	// Check that a load returns a copy of the input object
	b, ok := m.LoadAndDelete("k")
	assert.True(t, ok)
	assertDeepCopies(t, a, b)

	// Check that a load returns nil after a delete
	c, ok := m.Load("k")
	assert.False(t, ok)
	assert.Nil(t, c)

	// Now check the non-existing case
	d, ok := m.LoadAndDelete("k")
	assert.False(t, ok)
	assert.Nil(t, d)
}

func TestAgentMap_LoadOrStore(t *testing.T) {
	var m watchable.AgentMap

	a := &manager.AgentInfo{Name: "value"}
	m.Store("k", a)

	b := &manager.AgentInfo{Name: "value"}
	assertDeepCopies(t, a, b)

	c, ok := m.LoadOrStore("k", b)
	assert.True(t, ok)
	assertDeepCopies(t, a, c)
	assertDeepCopies(t, b, c)

	d, ok := m.LoadOrStore("k", b)
	assert.True(t, ok)
	assertDeepCopies(t, a, d)
	assertDeepCopies(t, b, d)
	assertDeepCopies(t, c, d)

	e, ok := m.LoadOrStore("x", a)
	assert.False(t, ok)
	assertDeepCopies(t, a, e)
	assertDeepCopies(t, b, e)
	assertDeepCopies(t, c, e)
	assertDeepCopies(t, d, e)
}

func TestAgentMap_Store(t *testing.T) {
	// TODO
}

func TestAgentMap_CompareAndSwap(t *testing.T) {
	// TODO
}

func TestAgentMap_Subscribe(t *testing.T) {
	ctx := dlog.NewTestContext(t, true)
	ctx, cancelCtx := context.WithCancel(ctx)
	var m watchable.AgentMap

	m.Store("a", &manager.AgentInfo{Name: "A"})
	m.Store("b", &manager.AgentInfo{Name: "B"})
	m.Store("c", &manager.AgentInfo{Name: "C"})

	ch := m.Subscribe(ctx)

	// Check that a complete snapshot is immediately available
	snapshot, ok := <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"a": &manager.AgentInfo{Name: "A"},
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
			},
			Updates: nil,
		},
		snapshot)

	// Check that writes don't block on the subscriber channel
	m.Store("d", &manager.AgentInfo{Name: "D"})
	m.Store("e", &manager.AgentInfo{Name: "E"})
	m.Store("f", &manager.AgentInfo{Name: "F"})

	// Check that those 3 updates get coalesced in to a single read
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"a": &manager.AgentInfo{Name: "A"},
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
				"d": &manager.AgentInfo{Name: "D"},
				"e": &manager.AgentInfo{Name: "E"},
				"f": &manager.AgentInfo{Name: "F"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "d", Value: &manager.AgentInfo{Name: "D"}},
				{Key: "e", Value: &manager.AgentInfo{Name: "E"}},
				{Key: "f", Value: &manager.AgentInfo{Name: "F"}},
			},
		},
		snapshot)

	// Check that deletes work
	m.Delete("a")
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
				"d": &manager.AgentInfo{Name: "D"},
				"e": &manager.AgentInfo{Name: "E"},
				"f": &manager.AgentInfo{Name: "F"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "a", Delete: true, Value: &manager.AgentInfo{Name: "A"}},
			},
		},
		snapshot)

	// Check that deletes work with LoadAndDlete
	m.LoadAndDelete("b")
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"c": &manager.AgentInfo{Name: "C"},
				"d": &manager.AgentInfo{Name: "D"},
				"e": &manager.AgentInfo{Name: "E"},
				"f": &manager.AgentInfo{Name: "F"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "b", Delete: true, Value: &manager.AgentInfo{Name: "B"}},
			},
		},
		snapshot)

	// Check that deletes coalesce with update
	m.Store("c", &manager.AgentInfo{Name: "c"})
	m.Delete("c")
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"d": &manager.AgentInfo{Name: "D"},
				"e": &manager.AgentInfo{Name: "E"},
				"f": &manager.AgentInfo{Name: "F"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "c", Value: &manager.AgentInfo{Name: "c"}},
				{Key: "c", Delete: true, Value: &manager.AgentInfo{Name: "c"}},
			},
		},
		snapshot)

	// Add some more writes, then close it
	m.Store("g", &manager.AgentInfo{Name: "G"})
	m.Store("h", &manager.AgentInfo{Name: "H"})
	m.Store("i", &manager.AgentInfo{Name: "I"})
	cancelCtx()
	// Because the 'close' happens asynchronously when the context ends, we need to wait a
	// moment to ensure that it's actually closed before we hit the next step.
	time.Sleep(20 * time.Millisecond)

	// Check that the writes get coalesced in to a "close".
	snapshot, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, snapshot)

	snapshot, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, snapshot)

	snapshot, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, snapshot)
}

func TestAgentMap_SubscribeSubset(t *testing.T) {
	ctx := dlog.NewTestContext(t, true)
	var m watchable.AgentMap

	m.Store("a", &manager.AgentInfo{Name: "A"})
	m.Store("b", &manager.AgentInfo{Name: "B"})
	m.Store("c", &manager.AgentInfo{Name: "C"})

	ch := m.SubscribeSubset(ctx, func(k string, v *manager.AgentInfo) bool {
		return v.Name != "ignoreme"
	})

	// Check that a complete snapshot is immediately available
	snapshot, ok := <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"a": &manager.AgentInfo{Name: "A"},
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
			},
		},
		snapshot)

	// Check that a no-op write doesn't trigger snapshot
	m.Store("a", &manager.AgentInfo{Name: "A"})
	select {
	case <-ch:
	case <-time.After(10 * time.Millisecond): // just long enough that we have confidence <-ch isn't going to happen
	}

	// Check that an overwrite triggers a new snapshot
	m.Store("a", &manager.AgentInfo{Name: "a"})
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"a": &manager.AgentInfo{Name: "a"},
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "a", Value: &manager.AgentInfo{Name: "a"}},
			},
		},
		snapshot)

	// Check that a now-ignored entry gets deleted from the snapshot
	m.Store("a", &manager.AgentInfo{Name: "ignoreme"})
	snapshot, ok = <-ch
	assert.True(t, ok)
	assertAgentMapSnapshotEqual(t,
		watchable.AgentMapSnapshot{
			State: map[string]*manager.AgentInfo{
				"b": &manager.AgentInfo{Name: "B"},
				"c": &manager.AgentInfo{Name: "C"},
			},
			Updates: []watchable.AgentMapUpdate{
				{Key: "a", Delete: true, Value: &manager.AgentInfo{Name: "a"}},
			},
		},
		snapshot)

	// Close the channel.  For sake of test coverage, let's do some things different than in the
	// non-Subset Subscribe test:
	//  1. Use m.Close() to close *all* channels, rather than canceling the Context to close
	//     just the one (not that more than one exists in this test)
	//  2. Don't have updates that will get coalesced in to the close.
	m.Close()
	snapshot, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, snapshot)

	// Now, since we've called m.Close(), let's check that subscriptions get already-closed
	// channels.
	ch = m.SubscribeSubset(ctx, func(k string, v *manager.AgentInfo) bool {
		return v.Name != "ignoreme"
	})
	snapshot, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, snapshot)
}
