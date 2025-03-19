package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	epoch        = int64(1672531200000) // Custom epoch (e.g., 2023-01-01 in ms)
	workerIDBits = 10
	sequenceBits = 12

	maxWorkerID = -1 ^ (-1 << workerIDBits) // 1023
	maxSequence = -1 ^ (-1 << sequenceBits) // 4095

	workerIDShift  = sequenceBits
	timestampShift = sequenceBits + workerIDBits
)

type Snowflake struct {
	mutex    sync.Mutex
	lastTS   int64
	sequence int64
	workerID int64
}

// NewSnowflake creates a new Snowflake generator with the given worker ID, returning an error for invalid IDs.
func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("workerID must be between 0 and %d", maxWorkerID)
	}
	return &Snowflake{
		workerID: workerID,
		lastTS:   -1,
		sequence: 0,
	}, nil
}

// NextID generates the next unique ID based on the snowflake algorithm, ensuring thread safety and time-based ordering.
func (s *Snowflake) NextID() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := currentTimeMillis()
	if now == s.lastTS {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// Sequence exhausted in this ms, wait for next millisecond
			for now <= s.lastTS {
				now = currentTimeMillis()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTS = now

	id := ((now - epoch) << timestampShift) |
		(s.workerID << workerIDShift) |
		s.sequence

	return id
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func main() {
	sf, err := NewSnowflake(1) // Worker ID = 1
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		id := sf.NextID()
		fmt.Printf("ID %d: %d\n", i+1, id)
	}
}
