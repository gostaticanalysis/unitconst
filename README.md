# unitconst

[![pkg.go.dev][gopkg-badge]][gopkg]

`unitconst` finds using a untyped constant without a unit.

```go
duration := 5 // 5 * time.Second is correct
time.Sleep(duration)
```

```sh
$ go vet -vettool=`which unitconst` -unitconst.type="time.Duration" main.go
./main.go:6:13: must not use a untyped constant without a unit
```

The default value of `unitconst.type` option is `time.Duration`.
`unitconst.type` accepts comma separated value such as `time.Duration, unit.Length`.

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/gostaticanalysis/unitconst
[gopkg-badge]: https://pkg.go.dev/badge/github.com/gostaticanalysis/unitconst?status.svg
