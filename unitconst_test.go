package unitconst_test

import (
	"testing"

	"github.com/gostaticanalysis/unitconst"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	defer unitconst.ExportSetFlagTypes("time.Duration, a.Length")()
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, unitconst.Analyzer, "a")
}
