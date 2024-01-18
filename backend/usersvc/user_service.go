package usersvc

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(context.Context, User) httputil.HttpError
	GetUser(ctx context.Context, id string) (User, httputil.HttpError)
	Authenticate(
		ctx context.Context,
		email, password string,
	) (token []byte, htterr httputil.HttpError)
	SearchUser(context.Context, UserSearchTerms) (User, httputil.HttpError)
}

type userService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user User) httputil.HttpError {
	httperr := s.repo.CreateUser(ctx, user)
	if httperr != nil && httputil.Is5xx(httperr.StatusCode()) {
		log.Error().Err(httperr).Msg("failed CreateUser")
	}
	return httperr
}

func (s *userService) GetUser(ctx context.Context, id string) (User, httputil.HttpError) {
	u, httperr := s.repo.GetUser(ctx, id)
	if httperr != nil {
		if httperr.StatusCode() >= 500 {
			log.Error().Err(httperr).Msg("failed GetUser")
		}
		return User{}, httperr
	}
	return u, nil
}

func (s *userService) SearchUser(
	ctx context.Context,
	st UserSearchTerms,
) (User, httputil.HttpError) {
	user, httperr := s.repo.SearchUser(ctx, st)
	if httperr != nil {
		if httperr.StatusCode() >= 500 {
			log.Error().Err(httperr).Msg("failed SearchUser")
		}
		return User{}, httperr
	}

	return user, nil
}

func (s *userService) Authenticate(
	ctx context.Context,
	email, password string,
) ([]byte, httputil.HttpError) {
	user, httperr := s.SearchUser(ctx, UserSearchTerms{Email: email})
	if httperr != nil {
		return nil, httperr
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, httputil.NewHttpError(http.StatusUnauthorized, "", "")
	}

	// TODO generate bearer token - for now just returning userId
	// h := sha256.New()
	return []byte(user.Id), nil

	// return h.Sum(nil), nil
}
