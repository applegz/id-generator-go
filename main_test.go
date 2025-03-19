package main

import (
	"testing"
)

func TestNewSnowflake(t *testing.T) {
	tests := []struct {
		name      string
		workerID  int64
		wantError bool
	}{
		{"validWorkerID", 0, false},
		{"maxWorkerID", maxWorkerID, false},
		{"negativeWorkerID", -1, true},
		{"exceedsMaxWorkerID", maxWorkerID + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snowflake, err := NewSnowflake(tt.workerID)

			if (err != nil) != tt.wantError {
				t.Errorf("unexpected error state; got err = %v, wantError = %v", err, tt.wantError)
			}

			if snowflake != nil && snowflake.workerID != tt.workerID {
				t.Errorf("unexpected workerID; got = %d, want = %d", snowflake.workerID, tt.workerID)
			}
		})
	}
}

func TestSnowflakeNextID(t *testing.T) {
	tests := []struct {
		name     string
		workerID int64
	}{
		{"basicGeneration", 1},
		{"zeroWorkerID", 0},
		{"maxWorkerID", maxWorkerID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf, err := NewSnowflake(tt.workerID)
			if err != nil {
				t.Fatalf("failed to create Snowflake: %v", err)
			}

			// Generate a few IDs and check for uniqueness
			ids := make(map[int64]struct{})
			for i := 0; i < 100; i++ {
				id := sf.NextID()
				if _, exists := ids[id]; exists {
					t.Errorf("duplicate ID generated: %d", id)
				} else {
					ids[id] = struct{}{}
				}
			}

			// Test rollover case
			sf.mutex.Lock()
			sf.sequence = maxSequence
			sf.lastTS = currentTimeMillis()
			sf.mutex.Unlock()

			id := sf.NextID()
			if id == 0 {
				t.Error("rollover ID incorrectly generated 0")
			}
		})
	}
}
