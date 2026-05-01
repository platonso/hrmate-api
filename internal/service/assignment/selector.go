package assignment

import (
	"math/rand"

	"github.com/google/uuid"
	errs "github.com/platonso/hrmate-api/internal/errors"
)

type HRWorkload struct {
	UserID            uuid.UUID
	PendingFormsCount int
}

func SelectOptimalHR(hrs []HRWorkload) (uuid.UUID, error) {
	if len(hrs) == 0 {
		return uuid.Nil, errs.ErrNoAvailableExecutors
	}

	minWorkload := hrs[0].PendingFormsCount
	candidates := []uuid.UUID{hrs[0].UserID}

	for _, hr := range hrs[1:] {
		switch {
		case hr.PendingFormsCount < minWorkload:
			minWorkload = hr.PendingFormsCount
			candidates = []uuid.UUID{hr.UserID}
		case hr.PendingFormsCount == minWorkload:
			candidates = append(candidates, hr.UserID)

		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	randomIndex := rand.Intn(len(candidates))
	return candidates[randomIndex], nil
}
