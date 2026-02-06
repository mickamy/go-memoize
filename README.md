# go-memoize

A generic function memoization library for Go with singleflight support.

## Features

- **Generics**: Type-safe memoization for `func(K) (V, error)`
- **Singleflight**: Deduplicates concurrent calls with the same key
- **TTL**: Optional cache expiration
- **Error handling**: Errors are not cached — failed calls are retried on next invocation
- **Zero dependencies**: Standard library only

## Install

```sh
go get github.com/mickamy/go-memoize
```

## Usage

### Simple API

```go
cachedGetUser := memoize.Do(func(id int) (User, error) {
    return db.GetUser(ctx, id)
}, memoize.WithTTL(5 * time.Minute))

u1, err := cachedGetUser(1) // executes function
u2, err := cachedGetUser(1) // returns cached result
```

### Full API

```go
userCache := memoize.New(func(id int) (User, error) {
    return db.GetUser(ctx, id)
}, memoize.WithTTL(5 * time.Minute))

u1, err := userCache.Get(1) // executes function
u2, err := userCache.Get(1) // returns cached result
userCache.Forget(1)          // invalidate a single key
userCache.Purge()            // clear all cache
```

### Multiple arguments

Use a struct as the key:

```go
type UserKey struct {
    OrgID  int
    UserID int
}

cachedGet := memoize.Do(func(k UserKey) (*User, error) {
    return db.GetUser(k.OrgID, k.UserID)
})
```

## API

| Function / Method    | Description                                    |
|----------------------|------------------------------------------------|
| `Do(fn, ...Option)`  | Wrap a function with memoization               |
| `New(fn, ...Option)` | Create a `Memo` instance with cache management |
| `Memo.Get(key)`      | Execute or return cached result                |
| `Memo.Forget(key)`   | Invalidate a single key                        |
| `Memo.Purge()`       | Clear all cache                                |
| `WithTTL(d)`         | Set cache expiration duration                  |

## License

[MIT](./LICENSE)