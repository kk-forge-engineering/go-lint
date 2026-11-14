package layers_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kk-forge-engineering/go-lint/v3/features/layers"
)

func TestAnalyzerReportsEachProblemOfTheFixture(t *testing.T) {
	// The test runs the analyzer under the name buildtag, so that analysistest checks `// want`
	// in files that the host build leaves out. analysistest checks those files only for a few
	// analyzer names, and the layers analyzer reports in them.
	analyzer := *layers.Analyzer
	analyzer.Name = "buildtag"
	analysistest.Run(t, analysistest.TestData()+"/app", &analyzer, "./...")
}
