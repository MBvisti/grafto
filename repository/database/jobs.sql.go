// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: jobs.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const deleteJob = `-- name: DeleteJob :exec
delete from jobs where id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteJob, id)
	return err
}

const failJob = `-- name: FailJob :exec
update jobs
    SET state = $1, updated_at = $2, scheduled_for = $3, failed_attempts = failed_attempts + 1
WHERE id = $4
`

type FailJobParams struct {
	State        int32
	UpdatedAt    time.Time
	ScheduledFor time.Time
	ID           uuid.UUID
}

func (q *Queries) FailJob(ctx context.Context, arg FailJobParams) error {
	_, err := q.db.Exec(ctx, failJob,
		arg.State,
		arg.UpdatedAt,
		arg.ScheduledFor,
		arg.ID,
	)
	return err
}

const insertJob = `-- name: InsertJob :exec
insert into jobs
    (id, created_at, updated_at, failed_attempts, state, instructions, scheduled_for, executor, repeatable_id)
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type InsertJobParams struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	FailedAttempts int32
	State          int32
	Instructions   pgtype.JSONB
	ScheduledFor   time.Time
	Executor       string
	RepeatableID   sql.NullString
}

func (q *Queries) InsertJob(ctx context.Context, arg InsertJobParams) error {
	_, err := q.db.Exec(ctx, insertJob,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.FailedAttempts,
		arg.State,
		arg.Instructions,
		arg.ScheduledFor,
		arg.Executor,
		arg.RepeatableID,
	)
	return err
}

const queryJobs = `-- name: QueryJobs :many
update jobs
    set state = $1, updated_at = $2
    where id in (
        select id
        from jobs as inner_jobs
        where inner_jobs.state = $4::int 
        and inner_jobs.scheduled_for::timestamptz <= $5::timestamptz 
        and inner_jobs.failed_attempts < $6::int
        order by inner_jobs.scheduled_for
        for update skip locked
        limit $3
    )
returning id, created_at, updated_at, scheduled_for, failed_attempts, state, instructions, executor, repeatable_id
`

type QueryJobsParams struct {
	State               int32
	UpdatedAt           time.Time
	Limit               int32
	InnerState          int32
	InnerScheduledFor   time.Time
	InnerFailedAttempts int32
}

func (q *Queries) QueryJobs(ctx context.Context, arg QueryJobsParams) ([]Job, error) {
	rows, err := q.db.Query(ctx, queryJobs,
		arg.State,
		arg.UpdatedAt,
		arg.Limit,
		arg.InnerState,
		arg.InnerScheduledFor,
		arg.InnerFailedAttempts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Job
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ScheduledFor,
			&i.FailedAttempts,
			&i.State,
			&i.Instructions,
			&i.Executor,
			&i.RepeatableID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const repeatableJobExists = `-- name: RepeatableJobExists :one
select exists(select 1 from jobs where repeatable_id = $1)
`

func (q *Queries) RepeatableJobExists(ctx context.Context, repeatableID sql.NullString) (bool, error) {
	row := q.db.QueryRow(ctx, repeatableJobExists, repeatableID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const rescheduleJob = `-- name: RescheduleJob :exec
update jobs
    set state = $1, updated_at = $2, scheduled_for  = $3, failed_attempts = 0
    where id = $4
`

type RescheduleJobParams struct {
	State        int32
	UpdatedAt    time.Time
	ScheduledFor time.Time
	ID           uuid.UUID
}

func (q *Queries) RescheduleJob(ctx context.Context, arg RescheduleJobParams) error {
	_, err := q.db.Exec(ctx, rescheduleJob,
		arg.State,
		arg.UpdatedAt,
		arg.ScheduledFor,
		arg.ID,
	)
	return err
}