package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestSendAndList(t *testing.T) {
	dispatcher := &dispatcherStub{}
	svc := notification.NewService(dispatcher)

	cmd, _ := assembler.FromRequest(dto.NotificationRequest{Type: "email", Target: "user@example.com", Message: "hello"})
	resp, err := svc.Send(context.Background(), cmd)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	list := svc.List(context.Background(), query.Options{Limit: 10})
	require.Len(t, list, 1)

	found, ok := svc.Get(context.Background(), resp.ID)
	require.True(t, ok)
	require.Equal(t, resp.ID, found.ID)
}

func TestSendDispatchError(t *testing.T) {
	dispatcher := &dispatcherStub{err: errors.New("fail")}
	svc := notification.NewService(dispatcher)
	cmd, _ := assembler.FromRequest(dto.NotificationRequest{Type: "email", Target: "x", Message: "y"})
	_, err := svc.Send(context.Background(), cmd)
	require.Error(t, err)
}

type dispatcherStub struct {
	err error
}

func (d *dispatcherStub) Dispatch(ctx context.Context, target, message string) error {
	return d.err
}
