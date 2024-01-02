package emailsvc

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailService interface {
	ThreadSearch(
		ctx context.Context,
		st ThreadSearchTerms,
	) (EmailThread, httputil.HttpError)
	AddEmail(
		ctx context.Context,
		threadId primitive.ObjectID,
		email Email,
	) httputil.HttpError
}

type emailService struct {
	repo EmailRepo
}

func NewEmailService(repo EmailRepo) *emailService {
	return &emailService{
		repo: repo,
	}
}

func (s *emailService) ThreadSearch(
	ctx context.Context,
	st ThreadSearchTerms,
) (EmailThread, httputil.HttpError) {
	if st == (ThreadSearchTerms{}) {
		return EmailThread{}, httputil.NewHttpError(
			http.StatusBadRequest,
			"missing ThreadSearchTerms",
			"",
		)
	}

	thread, err := s.repo.ThreadSearch(ctx, st)
	if err != nil {
		return EmailThread{}, err
	}

	return thread, nil
}

func (s *emailService) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email Email,
) httputil.HttpError {
	if threadId == primitive.NilObjectID {
		return httputil.NewHttpError(http.StatusBadRequest, "missing threadId", "")
	}
	if email == (Email{}) {
		return httputil.NewHttpError(http.StatusBadRequest, "missing email", "")
	}

	if err := s.repo.AddEmail(ctx, threadId, email); err != nil {
		return err
	}
	return nil
}
