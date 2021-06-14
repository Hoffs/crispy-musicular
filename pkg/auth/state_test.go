package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	st State
}

func (r *mockRepo) SetState(s State) error {
	r.st = s
	return nil
}

func (r *mockRepo) GetState() (State, error) {
	return r.st, nil
}

func (r *mockRepo) ClearState() error {
	r.st = State{}
	return nil
}

func TestGetStateEmpty(t *testing.T) {
	r := mockRepo{}
	s, err := NewService(&r)

	st, err := s.GetState()

	require.NoError(t, err)
	require.Equal(t, st, State{})
}

func TestGetStateNonEmpty(t *testing.T) {
	r := mockRepo{}
	s, err := NewService(&r)

	err = s.SetState(State{"Refresh", "User", "Drive", "Youtube"})
	st, err := s.GetState()

	require.NoError(t, err)
	require.Equal(t, st, State{RefreshToken: "Refresh", User: "User", DriveRefreshToken: "Drive", YoutubeRefreshToken: "Youtube"})
}

func TestSetStateNoValue(t *testing.T) {
	r := mockRepo{}
	s, err := NewService(&r)

	err = s.SetState(State{"Refresh", "", "Drive", "Youtube"})
	require.Error(t, err)

	err = s.SetState(State{"", "User", "Drive", "Youtube"})
	require.Error(t, err)
}

func TestSetStateValue(t *testing.T) {
	r := mockRepo{}
	s, err := NewService(&r)

	err = s.SetState(State{"Refresh", "User", "Drive", "Youtube"})
	require.NoError(t, err)

	err = s.SetState(State{"Refresh", "User", "", ""})
	require.NoError(t, err)
}

func TestClearState(t *testing.T) {
	r := mockRepo{}
	s, err := NewService(&r)

	err = s.SetState(State{"Refresh", "User", "Drive", "Youtube"})
	require.NoError(t, err)

	err = s.ClearState()
	require.NoError(t, err)

	st, err := s.GetState()
	require.Equal(t, st, State{})
}
