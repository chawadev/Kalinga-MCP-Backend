package google

import (
	"context"
	"fmt"

	"github.com/chawadev/kalinga-backend/internal/contacts"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/option"
)

type GoogleClient struct {
	service *people.Service
}

func NewClient(ctx context.Context, accessToken string) (*GoogleClient, error) {
	srv, err := people.NewService(ctx, option.WithTokenSource(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google People service: %w", err)
	}

	return &GoogleClient{service: srv}, nil
}

func (c *GoogleClient) FetchContacts(ctx context.Context, accessToken string) ([]contacts.Contact, error) {
	// TODO: Implement Google Contacts API integration
	// This would use the People API to fetch contacts
	return []contacts.Contact{}, nil
}

func (c *GoogleClient) CreateContact(ctx context.Context, contact *contacts.Contact) error {
	// TODO: Implement contact creation via Google People API
	return nil
}
