package queue

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInstructionsNotSet = errors.New("instructions not set")
	ErrExecutorNotSet     = errors.New("executor not set")
)

type Executor interface {
	process(ctx context.Context, msg []byte) error
	Name() string
}

type Job struct {
	id            uuid.UUID
	instructions  []byte
	executor      string
	failedAttemps int32
	scheduledFor  time.Time
}

func (j *Job) Schedule(t time.Time) *Job {
	j.scheduledFor = t

	return j
}

type newJobPayload struct {
	instructions []byte
	executor     string
}

func newJob(payload newJobPayload) (*Job, error) {
	if len(payload.instructions) == 0 {
		return nil, ErrInstructionsNotSet
	}

	if payload.executor == "" {
		return nil, ErrExecutorNotSet
	}

	job := &Job{
		id:           uuid.New(),
		instructions: payload.instructions,
		executor:     payload.executor,
		scheduledFor: time.Now().Add(1500 * time.Millisecond),
	}

	return job, nil
}
