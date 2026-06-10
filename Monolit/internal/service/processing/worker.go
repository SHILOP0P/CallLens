package processing

import (
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	defaultPollInterval = 2 * time.Second
	defaultWorkerLimit  = 10
	defaultRetryDelay   = 1 * time.Minute
	defaultStaleAfter   = 10 * time.Minute
)

type Worker struct {
	service      *Service
	pollInterval time.Duration
	workerLimit  int
	retryDelay   time.Duration
	staleAfter   time.Duration
	workerID     string
	log          logger.Logger
}

type WorkerOptions struct {
	PollInterval time.Duration
	Limit        int
	RetryDelay   time.Duration
	StaleAfter   time.Duration
}

func NewWorker(service *Service, opts WorkerOptions, log logger.Logger) *Worker {
	if log == nil {
		log = logger.NewNop()
	}

	if opts.PollInterval <= 0 {
		opts.PollInterval = defaultPollInterval
	}
	if opts.Limit <= 0 {
		opts.Limit = defaultWorkerLimit
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = defaultRetryDelay
	}
	if opts.StaleAfter <= 0 {
		opts.StaleAfter = defaultStaleAfter
	}

	return &Worker{
		service:      service,
		pollInterval: opts.PollInterval,
		workerLimit:  opts.Limit,
		retryDelay:   opts.RetryDelay,
		staleAfter:   opts.StaleAfter,
		workerID:     "processing-worker-" + uuid.NewString(),
		log:          log,
	}
}

func (w *Worker) Run(ctx context.Context) {
	if w.service == nil || w.service.processingJobRepository == nil {
		w.log.Warn(ctx, "processing worker skipped", zap.String("reason", "service_not_configured"))
		return
	}

	w.log.Info(
		ctx,
		"processing worker started",
		zap.Duration("poll_interval", w.pollInterval),
		zap.Int("worker_limit", w.workerLimit),
		zap.Duration("retry_delay", w.retryDelay),
		zap.Duration("stale_after", w.staleAfter),
		zap.String("worker_id", w.workerID),
	)

	for {
		select {
		case <-ctx.Done():
			w.log.Info(ctx, "processing worker stopped")
			return
		default:
		}

		if err := w.runBatch(ctx); err != nil {
			if ctx.Err() != nil {
				w.log.Info(ctx, "processing worker stopped")
				return
			}

			w.log.Error(ctx, "processing worker batch failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			w.log.Info(ctx, "processing worker stopped")
			return
		case <-time.After(w.pollInterval):
		}
	}
}

func (w *Worker) runBatch(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(w.workerLimit)

	claimed := 0

	for {
		job, err := w.service.processingJobRepository.TakeNext(groupCtx, w.workerID, w.staleAfter)
		if err != nil {
			if errors.Is(err, models.ErrNoProcessingJobs) {
				break
			}

			return err
		}

		claimed++
		claimedJob := job

		group.Go(func() error {
			if err := w.service.ProcessJob(groupCtx, claimedJob); err != nil {
				return w.handleJobError(groupCtx, claimedJob, err)
			}

			if _, err := w.service.processingJobRepository.MarkDone(groupCtx, claimedJob.ID); err != nil {
				return err
			}

			return nil
		})
	}

	if claimed == 0 {
		return nil
	}

	w.log.Info(ctx, "processing worker batch claimed jobs", zap.Int("count", claimed))

	return group.Wait()
}

func (w *Worker) handleJobError(ctx context.Context, job models.ProcessingJob, cause error) error {
	updatedJob, err := w.service.processingJobRepository.MarkRetry(ctx, job.ID, cause.Error(), w.retryDelay)
	if err != nil {
		return err
	}

	if updatedJob.Status == models.ProcessingJobStatusFailed {
		w.service.MarkJobFailed(ctx, job, cause)
	}

	w.log.Warn(
		ctx,
		"processing job failed",
		zap.String("job_id", job.ID.String()),
		zap.String("job_type", string(job.Type)),
		zap.String("entity_id", job.EntityUUID.String()),
		zap.String("status", string(updatedJob.Status)),
		zap.Int("attempts", updatedJob.Attempts),
		zap.Int("max_attempts", updatedJob.MaxAttempts),
		zap.Error(cause),
	)

	return nil
}
