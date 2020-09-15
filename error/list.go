package error

// List represent a list of errors
type List struct {
	errs []error
}

func (e *List) ReturnOrNil() error {
	if e.Empty() {
		return nil
	}
	return e
}

func (e *List) Empty() bool {
	return len(e.errs) == 0
}

func (e *List) Add(err error) {
	if err != nil {
		e.errs = append(e.errs, err)
	}
}

func (e *List) Error() string {
	errString := ""
	for _, err := range e.errs {
		errString += err.Error()
		errString += "\n"
	}
	return errString
}
