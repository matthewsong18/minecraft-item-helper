package main

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestPageStoreTraversalOwnershipAndExpiry(t *testing.T) {
	p := newPageStore()
	now := time.Now()
	p.now = func() time.Time { return now }
	token, e := p.create("alice", []string{"one", "two", "three"})
	if e != nil {
		t.Fatal(e)
	}
	if _, _, _, e = p.move(token, "bob", 1); !errors.Is(e, errPageOwner) {
		t.Fatalf("expected owner error, got %v", e)
	}
	for _, want := range []string{"two", "three", "three"} {
		got, index, total, e := p.move(token, "alice", 1)
		if e != nil || got != want || total != 3 || index > 2 {
			t.Fatalf("got %q page %d/%d err %v", got, index, total, e)
		}
	}
	got, index, _, e := p.move(token, "alice", -99)
	if e != nil || got != "one" || index != 0 {
		t.Fatalf("first page boundary got %q index %d err %v", got, index, e)
	}
	now = now.Add(14 * time.Minute)
	if _, _, _, e = p.move(token, "alice", 1); e != nil {
		t.Fatal(e)
	}
	now = now.Add(2 * time.Minute)
	if _, _, _, e = p.move(token, "alice", 1); !errors.Is(e, errPageExpired) {
		t.Fatalf("expected expiry, got %v", e)
	}
	if len(p.entries) != 0 {
		t.Fatalf("expired pages retained: %d", len(p.entries))
	}
}

func TestPageStoresAreIndependentUnderConcurrentRequests(t *testing.T) {
	p := newPageStore()
	first, err := p.create("alice", []string{"alice one", "alice two"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.create("bob", []string{"bob one", "bob two"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token, owner := first, "alice"
			if i%2 == 1 {
				token, owner = second, "bob"
			}
			_, _, _, moveErr := p.move(token, owner, 1)
			if moveErr != nil {
				t.Errorf("move failed: %v", moveErr)
			}
		}(i)
	}
	wg.Wait()
	alice, _, _, err := p.move(first, "alice", -1)
	if err != nil || alice != "alice one" {
		t.Fatalf("alice page crossed requests: %q, %v", alice, err)
	}
	bob, _, _, err := p.move(second, "bob", -1)
	if err != nil || bob != "bob one" {
		t.Fatalf("bob page crossed requests: %q, %v", bob, err)
	}
}
