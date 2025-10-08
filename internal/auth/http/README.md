# HTTP

## FAQ

### Why use DiscardUnknown: true?

When unmarshaling HTTP request bodies, we use:
```golang
protojson.UnmarshalOptions{DiscardUnknown: true}
```

This allows our API to remain forward-compatible. For example, consider the following:
1. New server version at /send now supports new field "time"
2. New client version now sends field "time"
3. New client can still interact with old servers - server will simply ignore "time" instead of failing to unmarshal.