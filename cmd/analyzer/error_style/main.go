package main

import (
	"github.com/MrEhbr/app/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.ErrStyleAnalyzer)
}
