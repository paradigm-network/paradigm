package errors

import "fmt"

type StoreErrType uint32

const (
	KeyNotFound StoreErrType = iota
	TooLate
	PassedIndex
	SkippedIndex
	NoRoot
	UnknownParticipant
)

type StoreErr struct {
	errType StoreErrType
	key     string
}

func NewStoreErr(errType StoreErrType, key string) StoreErr {
	return StoreErr{
		errType: errType,
		key:     key,
	}
}

func (e StoreErr) Error() string {
	m := ""
	switch e.errType {
	case KeyNotFound:
		m = "Not Found"
	case TooLate:
		m = "Too Late"
	case PassedIndex:
		m = "Passed Index"
	case SkippedIndex:
		m = "Skipped Index"
	case NoRoot:
		m = "No Root"
	case UnknownParticipant:
		m = "Unknown Participant"
	}

	return fmt.Sprintf("%s, %s", e.key, m)
}

func Is(err error, t StoreErrType) bool {
	storeErr, ok := err.(StoreErr)
	return ok && storeErr.errType == t
}


const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot() error
}


func NewDetailErr(err error, errcode ErrCode, errmsg string) DetailError {
	if err == nil {
		return nil
	}

	onterr, ok := err.(ontError)
	if !ok {
		onterr.root = err
		onterr.errmsg = err.Error()
		onterr.callstack = getCallStack(0, callStackDepth)
		onterr.code = errcode

	}
	if errmsg != "" {
		onterr.errmsg = errmsg + ": " + onterr.errmsg
	}

	return onterr
}
