package run

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// var (
// 	stateGauge = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "state_current",
// 			Help: "The current state",
// 		},
// 		[]string{"state"},
// 	)

// 	stateTimeGauge = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "state_time_seconds",
// 			Help: "The time spent in the current state",
// 		},
// 		[]string{"state"},
// 	)

//	stateChangesCounter = promauto.NewCounterVec(
//		prometheus.CounterOpts{
//			Name: "state_changes_total",
//			Help: "The number of state changes",
//		},
//		[]string{"from_state", "to_state"},
//	)
//
// )
type Metrics struct {
	currentState prometheus.Gauge
	timeInState  *prometheus.HistogramVec
	stateChanges *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{}
	m.currentState = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "automata_current_state",
		Help: "Current state of the automata",
	})
	m.timeInState = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "automata_time_in_state",
		Help: "Time spent in each state",
	}, []string{"state"})
	m.stateChanges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "automata_state_changes_total",
		Help: "Number of state changes",
	}, []string{"from_state", "to_state"})
	reg.MustRegister(m.currentState, m.timeInState, m.stateChanges)
	return m
}

var _ Runner = &LoopRunner{}

type Runner interface {
	Run(ctx context.Context, state states.AutomataState) error
}

func NewLoopRunner(logger *slog.Logger) *LoopRunner {
	logger = logger.With("subsystem", "StateRunner")
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	pMux.Handle("/metrics", promHandler)
	go func() {
		_ = http.ListenAndServe(":2112", pMux)
	}()

	return &LoopRunner{
		logger:  logger,
		metrics: m,
	}
}

type LoopRunner struct {
	logger  *slog.Logger
	metrics *Metrics
}

func (r *LoopRunner) Run(ctx context.Context, state states.AutomataState) error {
	startTime := time.Now()
	currentState := state.String()
	r.metrics.currentState.Set(1)
	for state != nil {
		r.logger.LogAttrs(ctx, slog.LevelInfo, "start running state", slog.String("state", state.String()))
		var err error
		select {
		case <-ctx.Done():
			r.logger.LogAttrs(ctx, slog.LevelInfo, "context cancelled, finishing")
			return nil
		default:
			state, err = state.Run(ctx)
			if err != nil {
				return fmt.Errorf("state %s run: %w", state.String(), err)
			}
			newState := state.String()
			r.metrics.stateChanges.WithLabelValues(currentState, newState).Inc() // Increase state changes counter
			r.metrics.currentState.Set(0)                                        // Reset previous state
			currentState = newState
			r.metrics.currentState.Set(1) // Set new state
			// Observing the time in the state in seconds
			r.metrics.timeInState.WithLabelValues(newState).Observe(time.Since(startTime).Seconds())
			startTime = time.Now()

		}
		r.metrics.currentState.Set(0) // Reset the last state gauge
		r.metrics.timeInState.WithLabelValues(currentState).Observe(time.Since(startTime).Seconds())

	}
	r.logger.LogAttrs(ctx, slog.LevelInfo, "no new state, finish")
	return nil
}
