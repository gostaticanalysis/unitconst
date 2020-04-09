package unitconst

func ExportSetFlagTypes(ts string) func() {
	tmp := flagTypes
	flagTypes = ts
	return func() {
		flagTypes = tmp
	}
}
