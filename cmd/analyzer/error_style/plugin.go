// This must be package main
package main

import (
	"github.com/MrEhbr/app/analyzer"
	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		analyzer.ErrStyleAnalyzer,
	}
}

var AnalyzerPlugin analyzerPlugin
