package itest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type trafficManager struct {
	NamespacePair
}

func WithTrafficManager(np NamespacePair, f func(ctx context.Context, ch NamespacePair)) {
	np.HarnessT().Run("Test_TrafficManager", func(t *testing.T) {
		ctx := WithT(np.HarnessContext(), t)
		require.NoError(t, np.GeneralError())
		th := &trafficManager{NamespacePair: np}
		th.PushHarness(ctx, th.setup, th.tearDown)
		defer th.PopHarness()
		f(ctx, th)
	})
}

func (th *trafficManager) setup(ctx context.Context) bool {
	t := getT(ctx)
	TelepresenceQuitOk(ctx)
	_, err := th.TelepresenceHelmInstall(ctx, false)
	return assert.NoError(t, err)
}

func (th *trafficManager) tearDown(ctx context.Context) {
	th.UninstallTrafficManager(ctx, th.ManagerNamespace())
}
