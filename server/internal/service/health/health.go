// Package health owns process liveness and external-dependency readiness.
// Liveness is dependency-free; readiness distinguishes hard requirements from
// optional capabilities that may temporarily degrade.
package health

import (
	"sync"
	"time"

	"jww/internal/domain"
	"jww/utils/logger"
)

const defaultLatencyWarnMS int64 = 1000

var log = logger.New().WithComponent("service.health")

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnavailable Status = "unavailable"
)

const (
	HealthStatusOK = "ok"

	ReadinessStatusReady    = "ready"
	ReadinessStatusDegraded = "degraded"
	ReadinessStatusNotReady = "not_ready"
)

type Checker interface {
	Check(domain.AppContext) error
}

type CheckFunc func(domain.AppContext) error

func (f CheckFunc) Check(ctx domain.AppContext) error { return f(ctx) }

type Dependency struct {
	Name     string
	Required bool
	Checker  Checker
}

type DependencyResult struct {
	Name     string
	Required bool
	Status   Status
	Latency  time.Duration
	Err      error
}

type HealthReport struct {
	Status string
}

type ReadinessReport struct {
	Status       string
	Ready        bool
	Dependencies []DependencyResult
}

type report struct {
	ready        bool
	degraded     bool
	dependencies []DependencyResult
}

type Service struct {
	config       Config
	dependencies []Dependency
}

func NewService(config Config, dependencies ...Dependency) *Service {
	return &Service{
		config:       config,
		dependencies: append([]Dependency(nil), dependencies...),
	}
}

// CheckHealth reports process liveness. It deliberately performs no dependency
// I/O, so a running process stays live even when an external service is down.
func (s *Service) CheckHealth(_ domain.AppContext) HealthReport {
	return HealthReport{Status: HealthStatusOK}
}

// CheckReadiness reports whether the instance should receive traffic. It is
// not ready only when a required dependency is unavailable; optional dependency
// failures leave it ready but degraded.
func (s *Service) CheckReadiness(ctx domain.AppContext) ReadinessReport {
	report := s.check(ctx)
	s.logWarnings(ctx, report.dependencies)
	status := ReadinessStatusReady
	switch {
	case !report.ready:
		status = ReadinessStatusNotReady
	case report.degraded:
		status = ReadinessStatusDegraded
	}
	return ReadinessReport{Status: status, Ready: report.ready, Dependencies: report.dependencies}
}

// logWarnings emits one warning per unavailable or slow dependency. Detailed
// status stays here and is never serialized to the HTTP client.
func (s *Service) logWarnings(ctx domain.AppContext, dependencies []DependencyResult) {
	threshold := s.config.latencyWarnThreshold()
	for _, dependency := range dependencies {
		attrs := []any{
			"dependency", dependency.Name,
			"required", dependency.Required,
			"latency_ms", dependency.Latency.Milliseconds(),
		}
		if dependency.Status == StatusUnavailable {
			if dependency.Err != nil {
				attrs = append(attrs, "error", dependency.Err.Error())
			}
			log.WithContext(ctx).Warn("readiness dependency unavailable", attrs...)
			continue
		}
		if dependency.Latency >= threshold {
			log.WithContext(ctx).Warn("readiness dependency slow", attrs...)
		}
	}
}

func (c Config) latencyWarnThreshold() time.Duration {
	ms := c.LatencyWarnMS
	if ms <= 0 {
		ms = defaultLatencyWarnMS
	}
	return time.Duration(ms) * time.Millisecond
}

// check evaluates every dependency concurrently and preserves constructor
// order in the report. Only failed required checks make ready false; failed
// optional checks mark the otherwise operational application as degraded.
func (s *Service) check(ctx domain.AppContext) report {
	results := make([]DependencyResult, len(s.dependencies))
	var wg sync.WaitGroup
	for i, dependency := range s.dependencies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startedAt := time.Now()
			result := DependencyResult{
				Name:     dependency.Name,
				Required: dependency.Required,
				Status:   StatusAvailable,
			}
			if dependency.Checker == nil {
				result.Status = StatusUnavailable
				result.Latency = time.Since(startedAt)
				results[i] = result
				return
			}
			if err := dependency.Checker.Check(ctx); err != nil {
				result.Status = StatusUnavailable
				result.Err = err
			}
			result.Latency = time.Since(startedAt)
			results[i] = result
		}()
	}
	wg.Wait()

	report := report{ready: true, dependencies: results}
	for _, result := range results {
		if result.Status == StatusAvailable {
			continue
		}
		if result.Required {
			report.ready = false
		} else {
			report.degraded = true
		}
	}
	return report
}
