## Go Notes

Lessons learned and Go-specific gotchas encountered while building this project.

### Slice Initialization: `make([]T, length)` vs `make([]T, 0, capacity)`

**Use `make([]int, 0, 10)` when appending:**
```go
s := make([]int, 0, 10)
for i := 0; i < 10; i++ {
    s = append(s, i)
}
```

**Use `make([]int, 10)` when using index assignment:**
```go
s := make([]int, 10)
for i := 0; i < 10; i++ {
    s[i] = i
}
```

---

*More notes will be added as I encounter interesting Go patterns and pitfalls.*