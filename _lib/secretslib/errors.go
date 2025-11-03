package secretslib

import "errors"

var ErrSecretNotFound = errors.New("secret not found")
var ErrProviderNotConfigured = errors.New("provider not configured")
