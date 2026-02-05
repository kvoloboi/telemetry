package node

import "sync/atomic"

// Counters holds the telemetry node metrics
type Counters struct {
	produced atomic.Int64
	dropped  atomic.Int64
	sent     atomic.Int64
	failed   atomic.Int64
}

func NewCounters() *Counters {
	return &Counters{}
}

func (c *Counters) IncProduced() { c.produced.Add(1) }
func (c *Counters) IncDropped()  { c.dropped.Add(1) }
func (c *Counters) IncSent()     { c.sent.Add(1) }
func (c *Counters) IncFailed()   { c.failed.Add(1) }

func (c *Counters) GetProduced() int64 { return c.produced.Load() }
func (c *Counters) GetDropped() int64  { return c.dropped.Load() }
func (c *Counters) GetSent() int64     { return c.sent.Load() }
func (c *Counters) GetFailed() int64   { return c.failed.Load() }
