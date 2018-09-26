package req

import (
	"github.com/paradigm-network/paradigm/network/actor"
)

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}
