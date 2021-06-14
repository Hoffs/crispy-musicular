package auth

import "errors"

type State struct {
	RefreshToken        string
	User                string
	DriveRefreshToken   string
	YoutubeRefreshToken string
}

func (s State) IsSet() bool {
	return s.RefreshToken != ""
}

type Service interface {
	GetState() (State, error)
	SetState(s State) error
	ClearState() error
}

type Repository interface {
	GetState() (State, error)
	SetState(s State) error
	ClearState() error
}

type service struct {
	r Repository
}

func NewService(r Repository) (s Service, err error) {
	if r == nil {
		err = errors.New("auth: repository is nil")
		return
	}

	return &service{r}, nil
}

func (s *service) GetState() (State, error) {
	// IMPROVEMENT: cache this to avoid querying DB everytime.
	return s.r.GetState()
}

func (s *service) SetState(st State) error {
	if st.RefreshToken == "" {
		return errors.New("state: RefreshToken must be not empty")
	}

	if st.User == "" {
		return errors.New("state: User must be not empty")
	}

	return s.r.SetState(st)
}

func (s *service) ClearState() error {
	return s.r.ClearState()
}
