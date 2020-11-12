package a

type Length int64

const (
	MilliMeter Length = 1
	CentiMeter Length = 10 * MilliMeter
	Meter      Length = 1000 * MilliMeter
	KiloMeter  Length = 1000 * Meter
)

func _() {
	var _ Length = 10             // want `must not use a untyped constant without a unit`
	var _ Length = 2 * MilliMeter // OK
	var _ Length = Length(2)      // OK
}
