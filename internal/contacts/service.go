package contacts

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
	googleClient *GoogleClient
}

func NewService(repo *Repository, googleClient *GoogleClient) *Service {
	return &Service{
		repo: repo,
		googleClient: googleClient,
	}
}

func (s *Service) CreateContact(ctx context.Context, userID string, req *CreateContactRequest) (*Contact, error) {
	contact := &Contact{
		UserID: userID,
		Name:   req.Name,
		Phone:  req.Phone,
		Email:  req.Email,
	}

	if err := s.repo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *Service) GetContacts(ctx context.Context, userID string) ([]Contact, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) SyncWithGoogle(ctx context.Context, userID string, accessToken string) ([]Contact, error) {
	if s.googleClient == nil {
		return nil, errors.New("Google client not configured")
	}

	googleContacts, err := s.googleClient.FetchContacts(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// TODO: Sync contacts with local database
	return googleContacts, nil
}
