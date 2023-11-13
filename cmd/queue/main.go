package main

import (
	"context"
	"os"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/repository/database"
)

func main() {
	ctx := context.Background()

	queuedJobsStream := make(chan []queue.Job)

	databaseConnection := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	defer databaseConnection.Close(ctx)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	db := database.New(databaseConnection)
	q := queue.New(db)

	emailJobExecutor := queue.NewEmailExecutor(&mailClient)
	executors := map[string]queue.Executor{
		emailJobExecutor.Name(): emailJobExecutor,
	}

	j, _ := queue.CreateEmailJob(queue.EmailInstructions{})
	q.Push(ctx, j)

	worker := queue.NewWorker(queuedJobsStream, db, executors)
	go worker.Start(ctx)

	q.Start(ctx, queuedJobsStream)
}