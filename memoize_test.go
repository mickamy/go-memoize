package memoize_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mickamy/go-memoize"
)

func TestMemo_Get(t *testing.T) {
	t.Parallel()

	t.Run("caches result", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32
		m := memoize.New(func(key string) (string, error) {
			callCount.Add(1)
			return "val:" + key, nil
		})

		v1, err := m.Get("a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v2, err := m.Get("a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v1 != "val:a" || v2 != "val:a" {
			t.Errorf("got %q, %q; want %q", v1, v2, "val:a")
		}
		if c := callCount.Load(); c != 1 {
			t.Errorf("function called %d times, want 1", c)
		}
	})

	t.Run("different keys cached independently", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32
		m := memoize.New(func(key string) (string, error) {
			callCount.Add(1)
			return "val:" + key, nil
		})

		v1, _ := m.Get("a")
		v2, _ := m.Get("b")

		if v1 != "val:a" {
			t.Errorf("got %q, want %q", v1, "val:a")
		}
		if v2 != "val:b" {
			t.Errorf("got %q, want %q", v2, "val:b")
		}
		if c := callCount.Load(); c != 2 {
			t.Errorf("function called %d times, want 2", c)
		}
	})

	t.Run("does not cache errors", func(t *testing.T) {
		t.Parallel()

		errFail := errors.New("fail")
		var callCount atomic.Int32
		m := memoize.New(func(key string) (string, error) {
			callCount.Add(1)
			if callCount.Load() == 1 {
				return "", errFail
			}
			return "ok", nil
		})

		_, err := m.Get("a")
		if !errors.Is(err, errFail) {
			t.Fatalf("got %v, want %v", err, errFail)
		}

		v, err := m.Get("a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "ok" {
			t.Errorf("got %q, want %q", v, "ok")
		}
		if c := callCount.Load(); c != 2 {
			t.Errorf("function called %d times, want 2", c)
		}
	})

	t.Run("TTL expires cache", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32
		m := memoize.New(func(key string) (string, error) {
			callCount.Add(1)
			return "val", nil
		}, memoize.WithTTL(50*time.Millisecond))

		m.Get("a")
		time.Sleep(100 * time.Millisecond)
		m.Get("a")

		if c := callCount.Load(); c != 2 {
			t.Errorf("function called %d times, want 2", c)
		}
	})

	t.Run("singleflight deduplicates concurrent calls", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32
		m := memoize.New(func(key string) (string, error) {
			callCount.Add(1)
			time.Sleep(10 * time.Millisecond)
			return "val", nil
		})

		const n = 100
		ready := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(n)

		for range n {
			go func() {
				defer wg.Done()
				<-ready
				v, err := m.Get("a")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if v != "val" {
					t.Errorf("got %q, want %q", v, "val")
				}
			}()
		}

		close(ready)
		wg.Wait()

		if c := callCount.Load(); c != 1 {
			t.Errorf("function called %d times, want 1", c)
		}
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		t.Parallel()

		m := memoize.New(func(key int) (int, error) {
			return key * 2, nil
		})

		var wg sync.WaitGroup
		for i := range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, err := m.Get(i % 10)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if v != (i%10)*2 {
					t.Errorf("got %d, want %d", v, (i%10)*2)
				}
			}()
		}

		wg.Wait()
	})
}

func TestMemo_Forget(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	m := memoize.New(func(key string) (string, error) {
		callCount.Add(1)
		return "val", nil
	})

	if _, err := m.Get("a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m.Forget("a")
	if _, err := m.Get("a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c := callCount.Load(); c != 2 {
		t.Errorf("function called %d times, want 2", c)
	}
}

func TestMemo_Purge(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	m := memoize.New(func(key string) (string, error) {
		callCount.Add(1)
		return "val", nil
	})

	if _, err := m.Get("a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := m.Get("b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m.Purge()
	if _, err := m.Get("a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := m.Get("b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c := callCount.Load(); c != 4 {
		t.Errorf("function called %d times, want 4", c)
	}
}

func TestDo(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	fn := memoize.Do(func(key string) (string, error) {
		callCount.Add(1)
		return "val:" + key, nil
	})

	v1, err := fn("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v2, err := fn("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v1 != "val:a" || v2 != "val:a" {
		t.Errorf("got %q, %q; want %q", v1, v2, "val:a")
	}
	if c := callCount.Load(); c != 1 {
		t.Errorf("function called %d times, want 1", c)
	}
}
