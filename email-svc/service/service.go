package service

import (
	"context"

	"github.com/benjamonnguyen/opendoor-chat-services/email-svc/model"
	"github.com/benjamonnguyen/opendoor-chat-services/email-svc/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IEmailService interface {
	ThreadSearch(
		ctx context.Context,
		st model.ThreadSearchTerms,
	) (*model.EmailThread, error)
	AddEmail(ctx context.Context, threadId primitive.ObjectID, email model.Email) error
}

type EmailService struct {
	EmailRepo repository.EmailRepo
}

func (s *EmailService) ThreadSearch(
	ctx context.Context,
	st model.ThreadSearchTerms,
) (*model.EmailThread, error) {
	if st == (model.ThreadSearchTerms{}) {
		return nil, status.Error(codes.InvalidArgument, "missing ThreadSearchTerms")
	}

	thread, err := s.EmailRepo.ThreadSearch(ctx, st)
	if err != nil {
		return nil, err
	}

	return thread, err
}

func (s *EmailService) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email model.Email,
) error {
	if threadId == primitive.NilObjectID {
		return status.Error(codes.InvalidArgument, "missing threadId")
	}
	if email == (model.Email{}) {
		return status.Error(codes.InvalidArgument, "missing email")
	}

	return s.EmailRepo.AddEmail(ctx, threadId, email)
}
