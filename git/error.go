package git

import (
	"time"
	"fmt"
)

type ErrExecTimeout struct {
	Duration time.Duration
}

func IsErrExecTimeout(err error) bool {
	_, ok := err.(ErrExecTimeout)
	return ok
}

func (err ErrExecTimeout) Error() string {
	return fmt.Sprintf("execution is timeout [duration: %v]", err.Duration)
}


