package naming_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kk-forge-engineering/go-lint/v3/features/naming"
)

func TestAnalyzerReportsEachProblemOfTheFixture(t *testing.T) {
	analysistest.Run(t, analysistest.TestData()+"/app", naming.Analyzer, "./...")
}
