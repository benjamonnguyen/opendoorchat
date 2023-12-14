package service

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/repo"
)

type UserService interface {
	GetUser(ctx context.Context, id string) (model.User, error)
}

type userService struct {
	repo repo.UserRepo
}

func NewUserService(repo repo.UserRepo) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetUser(ctx context.Context, id string) (model.User, httputil.HttpError) {
	if id == "" {
		return model.User{}, httputil.NewHttpError(
			http.StatusBadRequest,
			"required id is blank",
			"",
		)
	}

	u, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return model.User{}, httputil.HttpErrorFromErr(err)
	}
	return u, nil
}
