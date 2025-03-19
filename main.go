package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	epoch        = int64(1672531200000) // Custom epoch (2023-01-01 in ms)
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
	counter  int64 // Tracks generated IDs
}

func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("workerID must be between 0 and %d", maxWorkerID)
	}
	sf := &Snowflake{
		workerID: workerID,
		lastTS:   -1,
		sequence: 0,
	}
	go sf.logRate() // Start logging in background
	return sf, nil
}

func (s *Snowflake) NextID() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := currentTimeMillis()
	if now == s.lastTS {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.lastTS {
				now = currentTimeMillis()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTS = now
	s.counter++ // Increment count for logging

	id := ((now - epoch) << timestampShift) |
		(s.workerID << workerIDShift) |
		s.sequence

	return id
}

func (s *Snowflake) logRate() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		count := s.counter
		s.counter = 0 // Reset for next interval
		s.mutex.Unlock()

		fmt.Printf("[ID Stats] Generated %d IDs in the last 1 second\n", count)
	}
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func main() {
	sf, err := NewSnowflake(1)
	if err != nil {
		panic(err)
	}

	// Simulate high-throughput generation (example)
	for i := 0; i < 100000; i++ {
		_ = sf.NextID()
		// Optionally sleep for demo purposes
		time.Sleep(10 * time.Microsecond)
	}

	// Keep program running to see logging (for demo)
	select {}
}
