package event

import "time"

type FuturePubEventStorage interface {
	Add(event DomainEvent, pubTime time.Time) error
	PopLatestBeforeATime(theTime time.Time) (tag string, data []byte, err error)
	PopEarliestAfterATime(theTime time.Time) (tag string, data []byte, err error)
}
