package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	parser := failedConnEventParser{}

	t.Run("valid ssh login failed line", func(t *testing.T) {
		line := "Aug 23 03:20:21 ns356780 sshd[26573]: Invalid user trackmania from 51.145.141.8 port 49700"

		event, err := parser.Parse(line)
		require.NoError(t, err)
		require.Equal(t, "trackmania", event.Username)
		require.Equal(t, "51.145.141.8", event.IPAddress.String())
		require.Equal(t, 49700, event.Port)
		require.Equal(t, "trackmania", event.Username)
		require.Equal(t, time.Date(time.Now().Year(), 8, 23, 3, 20, 21, 0, time.UTC), event.Timestamp)
	})

	t.Run("valid ssh login failed line", func(t *testing.T) {
		line := "Aug 23 03:24:46 ns356780 sshd[26658]: input_userauth_request: invalid user sf [preauth]"

		_, err := parser.Parse(line)
		require.Error(t, err)
	})
}
