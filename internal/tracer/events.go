package tracer

import (
	"time"

	"crypto-tracker/internal/model"
)

// EventType identifies the kind of real-time progress update sent to the Bubbletea UI.
type EventType string

const (
	EventInit          EventType = "INIT"
	EventScanningNode  EventType = "SCANNING_NODE"
	EventHopDiscovered EventType = "HOP_DISCOVERED"
	EventOutputSkipped EventType = "OUTPUT_SKIPPED"
	EventExchangeHit   EventType = "EXCHANGE_HIT"
	EventComplete      EventType = "COMPLETE"
	EventError         EventType = "ERROR"
)

// TraceEvent is streamed from the BFS Tracing Engine to the Bubbletea UI loop.
type TraceEvent struct {
	Type        EventType
	Timestamp   time.Time
	Depth       int
	MaxDepth    int
	QueueSize   int
	VisitedCnt  int
	CurrentAddr string
	Message     string
	Step        *model.TraceStep
	Skipped     *model.SkippedOutput
	Result      *model.TraceResult
	Err         error
}

