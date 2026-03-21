package db

import (
	"context"
	"testing"
	"time"
)

func TestWithContext(t *testing.T) {
	d := &Database{}
	ctx := context.Background()
	
	newCtx, cancel := d.WithContext(ctx)
	if cancel == nil {
		t.Fatal("cancel function should not be nil")
	}
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("context should have a deadline")
	}

	expectedDeadline := time.Now().Add(DatabaseTimeout)
	if deadline.After(expectedDeadline.Add(time.Second)) || deadline.Before(expectedDeadline.Add(-time.Second)) {
		t.Errorf("deadline = %v, want near %v", deadline, expectedDeadline)
	}
}
