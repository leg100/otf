package watch

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	// input event channel
	in := make(chan otf.Event, 1)

	srv := &api{
		svc:    &fakeApp{ch: in},
		Logger: logr.Discard(),
	}

	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	// send one event and then close
	in <- otf.Event{
		Payload: &run.Run{},
		Type:    otf.EventRunCreated,
	}
	close(in)

	done := make(chan struct{})
	go func() {
		srv.watch(w, r)

		// expected output event
		want := `data: {"ID":"","CreatedAt":"0001-01-01T00:00:00Z","IsDestroy":false,"ForceCancelAvailableAt":null,"Message":"","Organization":"","Refresh":false,"RefreshOnly":false,"ReplaceAddrs":null,"PositionInQueue":0,"TargetAddrs":null,"AutoApply":false,"Speculative":false,"Status":"","StatusTimestamps":null,"WorkspaceID":"","ConfigurationVersionID":"","ExecutionMode":"","Plan":{"RunID":"","PhaseType":"","Status":"","StatusTimestamps":null},"Apply":{"RunID":"","PhaseType":"","Status":"","StatusTimestamps":null},"Latest":false,"Commit":null}
event: run_created

`
		assert.Equal(t, want, w.Body.String())

		done <- struct{}{}
	}()
	<-done
}
