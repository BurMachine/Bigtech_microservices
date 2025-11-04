package chat_repo

import "errors"

// Заглушки для ошибок в repo (можно использовать доменные, но для примера)
var (
	errRepoNotImplemented = errors.New("repository method not implemented")
	errRepoAlreadyExists  = errors.New("already exists") // Заглушка для ALREADY_EXISTS
	errRepoNotFound       = errors.New("not found")
	errRepoPermission     = errors.New("permission denied")
	errRepoInvalidArg     = errors.New("invalid argument")
)
