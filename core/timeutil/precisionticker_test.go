package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrecisionTicker(t *testing.T) {
	maxIterations := 100
	duration := 1 * time.Second

	start := time.Now()
	executions := 0
	ticker := NewPrecisionTicker(func() {
		executions++
	}, duration/time.Duration(maxIterations), WithMaxIterations(maxIterations))
	ticker.WaitForGracefulShutdown()
	end := time.Now()

	assert.InDelta(t, 1.0, float64(end.Sub(start).Nanoseconds())/float64(duration.Nanoseconds()), 0.05, "duration did not match expected value")
	assert.Equal(t, maxIterations, ticker.Iterations(), "iterations did not match expected value")
	assert.Equal(t, maxIterations, executions, "executions did not match expected value")
}
