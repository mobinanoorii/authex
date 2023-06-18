package web

import "time"

// Runtime used to track runtime data
type Runtime struct {
	Resolves    uint64
	Contexts    uint64
	Schemas     uint64
	Revocations uint64
	Credentials uint64
	StartTime   time.Time
	Version     string
}

func NewRuntime(version string) *Runtime {
	return &Runtime{
		StartTime: time.Now().UTC(),
		Version:   version,
	}
}
