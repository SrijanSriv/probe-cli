package vanillator_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-cli/v3/internal/engine/experiment/vanillator"
	"github.com/ooni/probe-cli/v3/internal/engine/mockable"
	"github.com/ooni/probe-cli/v3/internal/model"
	"golang.org/x/sys/execabs"
)

func TestRunWithExistingTor(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	path, err := execabs.LookPath("tor")
	if err != nil {
		t.Skip("there is no tor executable installed")
	}
	t.Log("found tor in path:", path)
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("using this tempdir", tempdir)
	m := vanillator.NewExperimentMeasurer(vanillator.Config{})
	ctx := context.Background()
	measurement := &model.Measurement{}
	callbacks := model.NewPrinterCallbacks(log.Log)
	sess := &mockable.Session{
		MockableLogger:  log.Log,
		MockableTempDir: tempdir,
	}
	if err = m.Run(ctx, sess, measurement, callbacks); err != nil {
		t.Fatal(err)
	}
}
