package iam

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
)

// OperationalLifetime is for time the token must be active.
type OperationalLifetime uint16

const (
	// RenewableToken is a token that can be renewed automatically on background as needed.
	// Current google policies fallback this to 1 hour.
	RenewableToken OperationalLifetime = 0
	// VeryShortTimedToken is valid for 5 minutes
	VeryShortTimedToken = RenewableToken + 300
	// ShortTimedToken is valid for 15 minutes
	ShortTimedToken = VeryShortTimedToken * 3
	// HalfTimedToken is valid for 30 minutes (or, half the default of a token)
	HalfTimedToken = ShortTimedToken * 2
	// NormalTimedToken is valid for 1 hour. This token will explicitly end after 1 hour and will not be
	// renewed on background
	NormalTimedToken = HalfTimedToken * 2
)

// GetImpersonateAccessToken calls GetImpersonateAccessTokenWithDelegates without delegates
func GetImpersonateAccessToken(ctx context.Context, destinationServiceAccountEmail string, lifetime OperationalLifetime, scopes ...string) (oauth2.TokenSource, error) {
	return GetImpersonateAccessTokenWithDelegates(ctx, destinationServiceAccountEmail, lifetime, scopes, nil)
}

// GetImpersonateAccessTokenWithDelegates is a QnD solution to get an impersonation access token
// for a destination service account identities and using delegates, if available. It returns
// an error in case that account couldn't be impersonated with the params given.
func GetImpersonateAccessTokenWithDelegates(ctx context.Context, destinationServiceAccountEmail string, lifetime OperationalLifetime, scopes []string, delegates []string) (oauth2.TokenSource, error) {
	if len(scopes) == 0 {
		return nil, errors.WithStack(errors.New("scope required"))
	}

	credentialsConfig := impersonate.CredentialsConfig{
		TargetPrincipal: destinationServiceAccountEmail,
		Scopes:          scopes,
	}

	// Only apply lifetime for non-renewable option
	if lifetime != RenewableToken {
		credentialsConfig.Lifetime = time.Second * time.Duration(lifetime)
	}

	if len(delegates) > 0 {
		credentialsConfig.Delegates = delegates
	}

	return impersonate.CredentialsTokenSource(ctx, credentialsConfig)
}
