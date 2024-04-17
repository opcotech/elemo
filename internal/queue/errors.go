package queue

import "errors"

var (
	ErrNoSchedule  = errors.New("no schedule set")        // no schedule set
	ErrNoTask      = errors.New("no task set")            // no task set
	ErrReceiveTask = errors.New("failed to receive task") // failed to receive task
	ErrSendTask    = errors.New("failed to send task")    // failed to send task
)
