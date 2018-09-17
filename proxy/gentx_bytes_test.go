package proxy

import (
	"testing"
	"time"
	"github.com/paradigm-network/paradigm/common/log"
)

func Test(t *testing.T) {

	var a = 0
	for i := 0; i < 10; i++ {
		a = a + 1
		tx := []byte(string(a))
		log.GetConsoleLogger("GenBytes").Info().Interface("tx", [][]byte{tx}).Msg("Transaction in [][]bytes")
		time.Sleep(time.Millisecond * 100)
	}

}
