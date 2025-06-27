package redis

import "errors"

type SynchroniserOption interface {
	Apply(*SynchroniserOptions) error
}

type SynchroniserOptions struct {
	DSN                string
	Password           string
	DB                 int
	LockTimeoutSeconds int // in seconds
}

type SynchroniserOptionFunc func(*SynchroniserOptions) error

func (f SynchroniserOptionFunc) Apply(opts *SynchroniserOptions) error {
	return f(opts)
}

func SynchroniserDB(db int) SynchroniserOption {
	return SynchroniserOptionFunc(func(opts *SynchroniserOptions) error {
		opts.DB = db
		return nil
	})
}

func SynchroniserDSN(dsn string) SynchroniserOption {
	return SynchroniserOptionFunc(func(opts *SynchroniserOptions) error {
		if dsn == "" {
			return errors.New("dsn cannot be empty")
		}
		opts.DSN = dsn
		return nil
	})
}

func SynchroniserPassword(password string) SynchroniserOption {
	return SynchroniserOptionFunc(func(opts *SynchroniserOptions) error {
		if password == "" {
			return errors.New("password cannot be empty")
		}
		opts.Password = password
		return nil
	})
}

func SynchroniserLockTimeOut(timeOutSeconds int) SynchroniserOption {
	return SynchroniserOptionFunc(func(opts *SynchroniserOptions) error {

		opts.LockTimeoutSeconds = timeOutSeconds
		return nil
	})
}
