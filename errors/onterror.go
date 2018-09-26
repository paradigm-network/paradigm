package errors

type ontError struct {
	errmsg    string
	callstack *CallStack
	root      error
	code      ErrCode
}

func (e ontError) Error() string {
	return e.errmsg
}

func (e ontError) GetErrCode() ErrCode {
	return e.code
}

func (e ontError) GetRoot() error {
	return e.root
}

func (e ontError) GetCallStack() *CallStack {
	return e.callstack
}
