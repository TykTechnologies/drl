package drl_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/drl"
)

// This is a black-box test suite for verifying the DRL

func TestDRLInit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ratelimiter := setupDRL(ctx)

	t.Run("add server", func(t *testing.T) {
		t.Parallel()

		server := drl.Server{
			HostName:   "127.0.0.1",
			ID:         "testing-node-id",
			LoadPerSec: 5,
		}
		err := addOrUpdateServer(ratelimiter, server)
		assert.NoError(t, err)
	})
}

func setupDRL(ctx context.Context) *drl.DRL {
	result := &drl.DRL{}
	result.Init(ctx)
	result.ThisServerID = "testing-node-id|127.0.0.1"
	return result
}

func addOrUpdateServer(dest *drl.DRL, server drl.Server) error {
	if dest == nil || !dest.Ready() {
		return errors.New("DRL not ready")
	}
	return dest.AddOrUpdateServer(server)
}
