package handler

import "github.com/google/uuid"

func storeNewUUID() (string, error) {
	return uuid.NewString(), nil
}
