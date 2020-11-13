package a

import (
	"context"
	"time"
)

func f() {
	var d1 time.Duration = 5 // want `must not use a untyped constant without a unit`
	time.Sleep(d1)

	d2 := 5 * time.Second // OK
	if true {
		d2 = 5 // want `must not use a untyped constant without a unit`
	}
	time.Sleep(d2)

	time.Sleep(5 * time.Second) // OK

	const i = 6
	time.Sleep(i)                              // want `must not use a untyped constant without a unit`
	time.Sleep(7)                              // want `must not use a untyped constant without a unit`
	time.Sleep(time.Duration(3600))            // OK
	time.Sleep(time.Duration(5) * time.Second) // OK
	time.Sleep(60 * 60)                        // want `must not use a untyped constant without a unit`
	time.Sleep(i * 60)                         // want `must not use a untyped constant without a unit`

	const c = 2i * 2i
	time.Sleep(10 + c) // want `must not use a untyped constant without a unit`

	const d3 = 7
	context.WithTimeout(context.Background(), d3) // want `must not use a untyped constant without a unit`

	(T{}).sleep(i) // want `must not use a untyped constant without a unit`

	_ = time.Duration(10) // OK

	const d4 = 8 * time.Second // OK
	time.Sleep(d4)

	time.Sleep(0) // OK
	const d5 = 0
	time.Sleep(d5) // want `must not use a untyped constant without a unit`
}

type T struct{}

func (T) sleep(d time.Duration) {
	time.Sleep(d)
}
