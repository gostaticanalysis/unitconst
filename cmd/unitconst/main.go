package main

import (
	"github.com/gostaticanalysis/unitconst"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(unitconst.Analyzer) }
