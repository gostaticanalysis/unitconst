package a

type Length int64

const (
	MilliMeter Length = 1
	CentiMeter Length = 10 * MilliMeter
	Meter      Length = 1000 * MilliMeter
	KiloMeter  Length = 1000 * Meter
)
