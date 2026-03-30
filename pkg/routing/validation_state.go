package routing

import "errors"

// ValidationState tracks configuration-time validation errors and whether
// invalid configuration should panic immediately.
type ValidationState struct {
	panicOnError bool
	errSink      *[]error
}

// NewValidationState constructs a validation state that panics on invalid
// configuration by default.
func NewValidationState() *ValidationState {
	errs := make([]error, 0)
	return &ValidationState{panicOnError: true, errSink: &errs}
}

// Clone returns a shallow clone that shares the same error sink while keeping
// an independent panic mode flag.
func (s *ValidationState) Clone() *ValidationState {
	if s == nil {
		return NewValidationState()
	}
	return &ValidationState{panicOnError: s.panicOnError, errSink: s.errSink}
}

// WithPanicOnError returns a clone with the provided panic behavior.
func (s *ValidationState) WithPanicOnError(enabled bool) *ValidationState {
	clone := s.Clone()
	clone.panicOnError = enabled
	return clone
}

// PanicOnError reports whether invalid configuration should panic.
func (s *ValidationState) PanicOnError() bool {
	if s == nil {
		return true
	}
	return s.panicOnError
}

// Handle records the error or panics immediately, depending on the current
// validation mode.
func (s *ValidationState) Handle(err error) {
	if err == nil {
		return
	}
	if s == nil || s.panicOnError {
		panic(err.Error())
	}
	if s.errSink == nil {
		errs := []error{err}
		s.errSink = &errs
		return
	}
	*s.errSink = append(*s.errSink, err)
}

// Errors returns a copy of all accumulated configuration errors.
func (s *ValidationState) Errors() []error {
	if s == nil || s.errSink == nil || len(*s.errSink) == 0 {
		return nil
	}
	errs := make([]error, len(*s.errSink))
	copy(errs, *s.errSink)
	return errs
}

// Err returns all accumulated configuration errors joined together.
func (s *ValidationState) Err() error {
	errs := s.Errors()
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
