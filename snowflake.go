package snowflake

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrStartTimeAhead = errors.New("start time is ahead of now")
	ErrOverTimeLimit  = errors.New("over the time limit")
)

const (
	period           = time.Millisecond
	BitLenSequence   = 12
	BitLenMachineID  = 10
	BitLenTime       = 41
	MaxValueSequence = 4096
)

type Snowflake struct {
	mx          sync.Mutex
	startTIme   int64
	elapsedTime int64
	idMachine   uint16
	sequence    uint16
}

type Option func(options *Snowflake) error

func WithStartTime(start time.Time) Option {
	return func(sf *Snowflake) error {
		if start.After(time.Now()) {
			return ErrStartTimeAhead
		}
		sf.startTIme = start.UnixMilli()
		return nil
	}
}

func New(IDmachine uint16, opts ...Option) (*Snowflake, error) {
	sf := new(Snowflake)

	for _, opt := range opts {
		err := opt(sf)
		if err != nil {
			return nil, err
		}
	}

	if sf.startTIme == 0 {
		sf.startTIme = time.Now().UnixMilli()
	}
	sf.idMachine = IDmachine

	return sf, nil
}

func (sf *Snowflake) NextID() (uint64, error) {
	sf.mx.Lock()
	defer sf.mx.Unlock()

	currentTime := time.Now().UnixMilli() - sf.startTIme
	if currentTime > sf.elapsedTime {
		sf.sequence = 0
		sf.elapsedTime = currentTime
	} else {
		sf.sequence++
		if sf.sequence == MaxValueSequence {
			sf.elapsedTime++
			sf.sequence = 0
			time.Sleep(period)
		}
	}

	return sf.getID()
}

func (sf *Snowflake) getID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, ErrOverTimeLimit
	}
	ID := uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.idMachine)<<BitLenSequence | uint64(sf.sequence)
	return ID, nil
}
