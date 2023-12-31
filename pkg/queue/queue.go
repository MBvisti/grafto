package queue

import (
	"context"
	"database/sql"
	"fmt"
	"hash"
	"hash/fnv"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

const (
	stateRunning = 1
	stateFailed  = 2
	stateQueued  = 3
	maxRetries   = 5
)

type queueStorage interface {
	QueryJobs(ctx context.Context, params database.QueryJobsParams) ([]database.Job, error)
	InsertJob(ctx context.Context, params database.InsertJobParams) error
	RepeatableJobExists(ctx context.Context, repeatableID sql.NullString) (bool, error)
}

type Queue struct {
	storage    queueStorage
	maxRetries int
	hasher     hash.Hash
}

func New(storage queueStorage) *Queue {
	hasher := fnv.New64a()

	return &Queue{
		storage,
		maxRetries,
		hasher,
	}
}

func (q *Queue) pull(ctx context.Context) ([]database.Job, error) {
	now := time.Now()

	return q.storage.QueryJobs(ctx, database.QueryJobsParams{
		State:               stateRunning,
		UpdatedAt:           database.ConvertToPGTimestamptz(now),
		Limit:               50,
		InnerState:          stateQueued,
		InnerScheduledFor:   database.ConvertToPGTimestamptz(now),
		InnerFailedAttempts: int32(q.maxRetries),
	})
}

func (q *Queue) Push(ctx context.Context, fn jobCreator) error {
	payload := fn.build()

	now := time.Now()

	return q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:           uuid.New(),
		CreatedAt:    database.ConvertToPGTimestamptz(now),
		UpdatedAt:    database.ConvertToPGTimestamptz(now),
		State:        stateQueued,
		Instructions: payload.instructions,
		Executor:     payload.executor,
		ScheduledFor: database.ConvertToPGTimestamptz(payload.scheduledFor),
	})
}

func (q *Queue) InitilizeRepeatingJobs(ctx context.Context, executors map[string]RepeatableExecutor) error {
	for name, executor := range executors {
		job, err := executor.generateJob()
		if err != nil {
			return err
		}

		q.hasher.Write(job.instructions)
		repeatJobID := fmt.Sprintf("%x", q.hasher.Sum(nil))

		if exists, err := q.storage.RepeatableJobExists(
			ctx, sql.NullString{String: repeatJobID, Valid: true}); err != nil {
			return err
		} else if exists {
			telemetry.Logger.Info("repeatable job already exists, skipping", "job", repeatJobID)
			return nil
		}

		now := time.Now()

		err = q.storage.InsertJob(ctx, database.InsertJobParams{
			ID:           uuid.New(),
			CreatedAt:    database.ConvertToPGTimestamptz(now),
			UpdatedAt:    database.ConvertToPGTimestamptz(now),
			ScheduledFor: database.ConvertToPGTimestamptz(job.scheduledFor),
			State:        stateQueued,
			Instructions: job.instructions,
			Executor:     name,
			RepeatableID: sql.NullString{String: repeatJobID, Valid: true},
		})
		if err != nil {
			telemetry.Logger.Error("failed to insert job", "error", err)
			return err
		}
	}

	return nil
}
