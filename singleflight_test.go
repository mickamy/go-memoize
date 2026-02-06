package memoize_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mickamy/go-memoize"
)

func TestSingleflight(t *testing.T) {
	t.Parallel()

	t.Run("deduplicates concurrent calls", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32
		sf := memoize.NewSingleflight[string, string]()

		const n = 100
		ready := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(n)

		for range n {
			go func() {
				defer wg.Done()
				<-ready
				val, err := sf.Do("key", func() (string, error) {
					callCount.Add(1)
					time.Sleep(10 * time.Millisecond)
					return "result", nil
				})
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if val != "result" {
					t.Errorf("got %q, want %q", val, "result")
				}
			}()
		}

		close(ready)
		wg.Wait()

		if count := callCount.Load(); count != 1 {
			t.Errorf("function called %d times, want 1", count)
		}
	})

	t.Run("propagates error to all callers", func(t *testing.T) {
		t.Parallel()

		errFailed := errors.New("failed")
		sf := memoize.NewSingleflight[string, string]()

		const n = 100
		ready := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(n)

		for range n {
			go func() {
				defer wg.Done()
				<-ready
				_, err := sf.Do("key", func() (string, error) {
					time.Sleep(10 * time.Millisecond)
					return "", errFailed
				})
				if !errors.Is(err, errFailed) {
					t.Errorf("got error %v, want %v", err, errFailed)
				}
			}()
		}

		close(ready)
		wg.Wait()
	})

	t.Run("different keys run independently", func(t *testing.T) {
		t.Parallel()

		var count1, count2 atomic.Int32
		sf := memoize.NewSingleflight[string, int]()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			if _, err := sf.Do("a", func() (int, error) {
				count1.Add(1)
				return 1, nil
			}); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := sf.Do("b", func() (int, error) {
				count2.Add(1)
				return 2, nil
			}); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()

		wg.Wait()

		if c := count1.Load(); c != 1 {
			t.Errorf("key 'a' called %d times, want 1", c)
		}
		if c := count2.Load(); c != 1 {
			t.Errorf("key 'b' called %d times, want 1", c)
		}
	})
}
