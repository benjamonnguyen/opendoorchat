package service

import (
	"context"
	"crypto/sha256"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoorchat/user-svc/model"
	"github.com/benjamonnguyen/opendoorchat/user-svc/repo"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(context.Context, model.User) httputil.HttpError
	GetUser(ctx context.Context, id string) (model.User, httputil.HttpError)
	Authenticate(
		ctx context.Context,
		email, password string,
	) (token []byte, htterr httputil.HttpError)
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

func (s *userService) CreateUser(ctx context.Context, user model.User) httputil.HttpError {
	httperr := s.repo.CreateUser(ctx, user)
	if httperr != nil && httputil.Is5xx(httperr.StatusCode()) {
		log.Error().Err(httperr).Msg("failed CreateUser")
	}
	return httperr
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
) ([]byte, httputil.HttpError) {
	user, httperr := s.SearchUser(ctx, model.UserSearchTerms{Email: email})
	if httperr != nil {
		return nil, httperr
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, httputil.NewHttpError(http.StatusUnauthorized, "", "")
	}

	// TODO generate bearer token
	h := sha256.New()
	if _, err := h.Write([]byte(email)); err != nil {
		return nil, httputil.HttpErrorFromErr(err)
	}

	return h.Sum(nil), nil
}
