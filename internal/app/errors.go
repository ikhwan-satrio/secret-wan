package app

// UserError is an error whose message is safe to show directly to a CLI user.
// Short is used as the error heading; Detail is the explanatory line.
type UserError struct {
	Short  string
	Detail string
}

func (e *UserError) Error() string {
	return e.Detail
}

func newUserError(short, detail string) error {
	return &UserError{Short: short, Detail: detail}
}

// RecipientChangeError reports that a recipient-list change could not be paired
// with a successful rekey. Restored says whether the previous list was written
// back successfully.
type RecipientChangeError struct {
	Err         error
	Restored    bool
	RollbackErr error
}

func (e *RecipientChangeError) Error() string {
	return e.Err.Error()
}

func (e *RecipientChangeError) Unwrap() error {
	return e.Err
}
