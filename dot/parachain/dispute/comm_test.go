package dispute

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendMessage(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		ch := make(chan any, 1)
		defer close(ch)

		err := sendMessage(ch, "test")
		require.NoError(t, err)
	})
	t.Run("timeout", func(t *testing.T) {
		ch := make(chan any)
		defer close(ch)

		err := sendMessage(ch, "test")
		require.NoError(t, err)
	})
}

func TestCall(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		receiver := make(chan any)
		response := make(chan any)
		defer close(receiver)
		defer close(response)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			select {
			case <-receiver:
				response <- "pong"
			case <-ctx.Done():
				require.NoError(t, ctx.Err())
			}
		}()

		res, err := call(receiver, "ping", response)
		require.NoError(t, err)
		require.Equal(t, "pong", res)
	})
	t.Run("timeout", func(t *testing.T) {
		receiver := make(chan any)
		response := make(chan any)
		defer close(receiver)
		defer close(response)

		res, err := call(receiver, "ping", response)
		require.Error(t, err)
		require.Nil(t, res)
	})
}
