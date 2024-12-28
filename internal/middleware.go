package internal

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/metrics"
	"github.com/sirupsen/logrus"
	"testgenerate_backend_subject/internal/app"
	"time"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	next   Service
	logger *logrus.Logger
}

func LoggingMiddleware(logger *logrus.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

func (mw loggingMiddleware) GetSubjects(ctx context.Context) (subjects []app.Subject, err error) {
	defer func(begin time.Time) {
		mw.logger.WithFields(logrus.Fields{
			"took":  time.Since(begin).Milliseconds(),
			"error": err,
		}).Info("method == GetSubjects")
	}(time.Now())
	return mw.next.GetSubjects(ctx)
}

func (mw loggingMiddleware) AddSubject(ctx context.Context, subjectAdd app.Subject) (err error) {
	defer func(begin time.Time) {
		mw.logger.WithFields(logrus.Fields{
			"took":  time.Since(begin).Milliseconds(),
			"error": err,
		}).Info("method == AddSubject")
	}(time.Now())
	return mw.next.AddSubject(ctx, subjectAdd)
}

func (mw loggingMiddleware) UpdateSubject(ctx context.Context, subject app.Subject) (err error) {
	defer func(begin time.Time) {
		mw.logger.WithFields(logrus.Fields{
			"took":  time.Since(begin).Milliseconds(),
			"error": err,
		}).Info("method == UpdateSubject")
	}(time.Now())
	return mw.next.UpdateSubject(ctx, subject)
}

func (mw loggingMiddleware) DeleteSubject(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.WithFields(logrus.Fields{
			"took":  time.Since(begin).Milliseconds(),
			"error": err,
		}).Info("method == DeleteSubject")
	}(time.Now())
	return mw.next.DeleteSubject(ctx, id)
}

// ----------------------------------------------------------------------------------------------------------------------
type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func InstrumentingMiddleware(requestCount metrics.Counter, requestLatency metrics.Histogram) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{
			requestCount:   requestCount,
			requestLatency: requestLatency,
			next:           next,
		}
	}
}

func (im instrumentingMiddleware) GetSubjects(ctx context.Context) (subjects []app.Subject, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "getSubjects", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	subjects, err = im.next.GetSubjects(ctx)
	return
}

func (im instrumentingMiddleware) AddSubject(ctx context.Context, subjectAdd app.Subject) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "addSubject", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	err = im.next.AddSubject(ctx, subjectAdd)
	return
}

func (im instrumentingMiddleware) UpdateSubject(ctx context.Context, subject app.Subject) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "updateSubject", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	err = im.next.UpdateSubject(ctx, subject)
	return
}

func (im instrumentingMiddleware) DeleteSubject(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "deleteSubject", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	err = im.next.DeleteSubject(ctx, id)
	return
}
