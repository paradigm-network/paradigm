package actor

import "github.com/paradigm-network/paradigm/network/actor/mailbox"

var (
	defaultDispatcher = mailbox.NewDefaultDispatcher(300)
)

var defaultMailboxProducer = mailbox.Unbounded()
