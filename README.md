# unitconst

[![godoc.org][godoc-badge]][godoc]

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
[godoc]: https://godoc.org/github.com/gostaticanalysis/unitconst
[godoc-badge]: https://img.shields.io/badge/godoc-reference-4F73B3.svg?style=flat-square&label=%20godoc.org

