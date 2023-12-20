package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/repo"
)

type UserService interface {
	GetUser(ctx context.Context, id string) (model.User, httputil.HttpError)
	Authenticate(
		ctx context.Context,
		email, password string,
	) (token string, htterr httputil.HttpError)
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

func (s *userService) Authenticate(
	ctx context.Context,
	email, password string,
) (string, httputil.HttpError) {
	user, httperr := s.repo.SearchUser(ctx, model.UserSearchTerms{Email: email})
	if httperr != nil {
		return "", httperr
	}
	devlog.Printf("userservice.Authenticate: SearchUser: %#v\n", user)

	if password == user.Password {
		// TODO generate bearer token
		h := sha256.New()
		if _, err := h.Write([]byte(email)); err != nil {
			return "", httputil.HttpErrorFromErr(err)
		}

		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}

	return "", httputil.NewHttpError(http.StatusUnauthorized, "", "")
}
