package service

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoorchat/email-svc/model"
	"github.com/benjamonnguyen/opendoorchat/email-svc/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailService interface {
	ThreadSearch(
		ctx context.Context,
		st model.ThreadSearchTerms,
	) (model.EmailThread, httputil.HttpError)
	AddEmail(ctx context.Context, threadId primitive.ObjectID, email model.Email) httputil.HttpError
}

type emailService struct {
	repo repo.EmailRepo
}

func NewEmailService(repo repo.EmailRepo) *emailService {
	return &emailService{
		repo: repo,
	}
}

func (s *emailService) ThreadSearch(
	ctx context.Context,
	st model.ThreadSearchTerms,
) (model.EmailThread, httputil.HttpError) {
	if st == (model.ThreadSearchTerms{}) {
		return model.EmailThread{}, httputil.NewHttpError(
			http.StatusBadRequest,
			"missing ThreadSearchTerms",
			"",
		)
	}

	thread, err := s.repo.ThreadSearch(ctx, st)
	if err != nil {
		return model.EmailThread{}, err
	}

	return thread, nil
}

func (s *emailService) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email model.Email,
) httputil.HttpError {
	if threadId == primitive.NilObjectID {
		return httputil.NewHttpError(http.StatusBadRequest, "missing threadId", "")
	}
	if email == (model.Email{}) {
		return httputil.NewHttpError(http.StatusBadRequest, "missing email", "")
	}

	if err := s.repo.AddEmail(ctx, threadId, email); err != nil {
		return err
	}
	return nil
}
