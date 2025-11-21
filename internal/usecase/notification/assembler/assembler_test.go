package assembler

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

func TestFromRequest(t *testing.T) {
	req := dto.NotificationRequest{Type: "email", Target: "user@example.com", Message: "hello"}
	cmd, err := FromRequest(req)
	require.NoError(t, err)
	require.Equal(t, req.Target, cmd.Target)

	_, err = FromRequest(dto.NotificationRequest{Type: "email", Target: "", Message: "hello"})
	require.Error(t, err)

	_, err = FromRequest(dto.NotificationRequest{Type: "email", Target: "x", Message: ""})
	require.Error(t, err)
}

