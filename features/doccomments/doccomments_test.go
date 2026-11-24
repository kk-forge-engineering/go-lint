package doccomments_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kk-forge-engineering/go-lint/v6/features/doccomments"
)

func TestAnalyzerReportsEachBlockInADocComment(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), doccomments.New(doccomments.Options{}).Analyzer(), "example")
}
