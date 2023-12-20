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
	"github.com/rs/zerolog/log"
)

type UserService interface {
	CreateUser(context.Context, model.User) error
	GetUser(ctx context.Context, id string) (model.User, httputil.HttpError)
	Authenticate(
		ctx context.Context,
		email, password string,
	) (token string, htterr httputil.HttpError)
	SearchUser(context.Context, model.UserSearchTerms) (model.User, httputil.HttpError)
}

type userService struct {
	repo repo.UserRepo
}

func NewUserService(repo repo.UserRepo) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user model.User) error {
	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("failed CreateUser")
	}
	return err
}

func (s *userService) GetUser(ctx context.Context, id string) (model.User, httputil.HttpError) {
	u, httperr := s.repo.GetUser(ctx, id)
	if httperr != nil {
		if httperr.StatusCode() >= 500 {
			log.Error().Err(httperr).Msg("failed GetUser")
		}
		return model.User{}, httperr
	}
	return u, nil
}

func (s *userService) SearchUser(
	ctx context.Context,
	st model.UserSearchTerms,
) (model.User, httputil.HttpError) {
	user, httperr := s.repo.SearchUser(ctx, st)
	if httperr != nil {
		if httperr.StatusCode() >= 500 {
			log.Error().Err(httperr).Msg("failed SearchUser")
		}
		return model.User{}, httperr
	}

	return user, nil
}

func (s *userService) Authenticate(
	ctx context.Context,
	email, password string,
) (string, httputil.HttpError) {
	user, httperr := s.SearchUser(ctx, model.UserSearchTerms{Email: email})
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
