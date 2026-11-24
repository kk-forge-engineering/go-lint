package layers_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kk-forge-engineering/go-lint/v6/features/layers"
)

func TestAnalyzerReportsEachProblemOfTheFixture(t *testing.T) {
	// The test runs the analyzer under the name buildtag, so that analysistest checks `// want`
	// in files that the host build leaves out. analysistest checks those files only for a few
	// analyzer names, and the layers analyzer reports in them.
	analyzer := *layers.New(layers.Options{}).Analyzer()
	analyzer.Name = "buildtag"
	analysistest.Run(t, analysistest.TestData()+"/app", &analyzer, "./...")
}

func TestAnalyzerReportsEachProblemBetweenTheAppsAndPackagesOfARepository(t *testing.T) {
	analyzer := *layers.New(layers.Options{}).Analyzer()
	analyzer.Name = "buildtag"
	for _, module := range []string{"repo/apps/api", "repo/apps/web", "repo/packages/logger", "standalone/packages/lib"} {
		analysistest.Run(t, analysistest.TestData()+"/"+module, &analyzer, "./...")
	}
}

func TestImportsKeyChangesWithTheModulesOfTheRepository(t *testing.T) {
	useCases := layers.New(layers.Options{})
	withAPI, err := useCases.ImportsKey(analysistest.TestData() + "/repo/apps/web")
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := os.CopyFS(repo, os.DirFS(analysistest.TestData()+"/repo")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(repo, "apps", "api")); err != nil {
		t.Fatal(err)
	}
	withoutAPI, err := useCases.ImportsKey(filepath.Join(repo, "apps", "web"))
	if err != nil {
		t.Fatal(err)
	}
	if withAPI == withoutAPI {
		t.Errorf("ImportsKey gave %s with and without the api app", withAPI)
	}
}
