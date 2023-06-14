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

// keep the list of participants and if they are active
var participants map[string]bool = map[string]bool{
	"0xe2Fb069045dFB19f3DD2B95A5A09D6F62984932d": true,
	"0x1e9Ee7293bc304A10a0b33D0FCCBDFF78463bE5c": true,
	"0x63791eb05F38Fdb8A34e4D70C4A8C75d671499b5": true,
}
