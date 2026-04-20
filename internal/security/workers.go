package security

import (
	"context"

	"github.com/riverqueue/river"
)

type PostActivityJobArgs struct {
	Activity any    `json:"activity"`
	Inbox    string `json:"inbox"`
	Actor    string `json:"actor"`
}

func (PostActivityJobArgs) Kind() string {
	return "post_activity_job_args"
}

type ActivityPosterWorker struct {
	river.WorkerDefaults[PostActivityJobArgs]
	Security *Security
}

func (w *ActivityPosterWorker) Work(ctx context.Context, job *river.Job[PostActivityJobArgs]) error {
	return w.Security.PostSigned(ctx, job.Args.Inbox, job.Args.Activity, job.Args.Actor)
}

func (s *Security) RegisterWorkers(workers *river.Workers) {
	worker := ActivityPosterWorker{
		Security: s,
	}

	river.AddWorker(workers, &worker)
}
