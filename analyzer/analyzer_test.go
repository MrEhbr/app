package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	tests := []string{"a"}

	analysistest.Run(t, testdata, ErrStyleAnalyzer, tests...)
}
