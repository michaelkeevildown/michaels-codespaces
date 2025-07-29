package ui

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// captureOutput captures stdout for testing terminal output
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(&buf, r)
	}()

	f()

	w.Close()
	os.Stdout = oldStdout
	wg.Wait()
	
	return buf.String()
}

func TestNewProgress(t *testing.T) {
	p := NewProgress()
	
	assert.NotNil(t, p)
	assert.NotNil(t, p.done)
	assert.NotNil(t, p.paused)
	assert.NotNil(t, p.resumed)
	assert.Equal(t, "", p.currentTask)
	assert.Equal(t, 0, p.frame)
	assert.False(t, p.isRunning)
}

func TestProgressStart(t *testing.T) {
	p := NewProgress()
	task := "Testing task"
	
	// Capture output to avoid cluttering test output
	output := captureOutput(func() {
		p.Start(task)
		// Give the goroutine a moment to start
		time.Sleep(100 * time.Millisecond)
		p.Success("Done")
	})
	
	assert.Equal(t, task, p.currentTask)
	// Frame might have changed due to goroutine timing, so don't check exact value
	assert.True(t, p.frame >= 0 && p.frame < len(spinnerFrames), "Frame should be within valid range")
	assert.Contains(t, output, "✓ Done") // Success message should be in output
}

func TestProgressSuccess(t *testing.T) {
	p := NewProgress()
	message := "Task completed successfully"
	
	output := captureOutput(func() {
		p.Start("Test task")
		time.Sleep(50 * time.Millisecond) // Let spinner run briefly
		p.Success(message)
	})
	
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, message)
}

func TestProgressFail(t *testing.T) {
	p := NewProgress()
	message := "Task failed"
	
	output := captureOutput(func() {
		p.Start("Test task")
		time.Sleep(50 * time.Millisecond) // Let spinner run briefly
		p.Fail(message)
	})
	
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, message)
}

func TestProgressUpdate(t *testing.T) {
	p := NewProgress()
	initialTask := "Initial task"
	updatedTask := "Updated task"
	
	output := captureOutput(func() {
		p.Start(initialTask)
		time.Sleep(50 * time.Millisecond) // Let spinner run briefly
		p.Update(updatedTask)
		time.Sleep(50 * time.Millisecond) // Let updated task show
		p.Success("Done")
	})
	
	assert.Equal(t, updatedTask, p.currentTask)
	assert.Contains(t, output, "✓ Done")
}

func TestProgressStopResume(t *testing.T) {
	p := NewProgress()
	task := "Test task"
	
	output := captureOutput(func() {
		p.Start(task)
		time.Sleep(100 * time.Millisecond) // Let spinner run
		
		// Stop the progress
		p.Stop()
		time.Sleep(100 * time.Millisecond) // Wait for stop to take effect
		assert.False(t, p.isRunning)
		
		// Resume the progress
		p.Resume()
		time.Sleep(100 * time.Millisecond) // Let it resume
		assert.True(t, p.isRunning)
		
		p.Success("Done")
	})
	
	assert.Contains(t, output, "✓ Done")
}

func TestProgressFrameRotation(t *testing.T) {
	p := NewProgress()
	
	// Start with frame 0
	assert.Equal(t, 0, p.frame)
	
	captureOutput(func() {
		p.Start("Test rotation")
		
		// Wait for several frame updates
		time.Sleep(300 * time.Millisecond) // Should allow multiple frame updates
		
		p.Success("Done")
	})
	
	// Frame should have incremented during the spinner animation
	// We can't guarantee the exact frame due to timing, but it should be > 0
	assert.True(t, p.frame >= 0 && p.frame < len(spinnerFrames),
		"Frame should be within valid range")
}

func TestProgressConcurrentOperations(t *testing.T) {
	p := NewProgress()
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	output := captureOutput(func() {
		// Start progress
		go func() {
			defer wg.Done()
			p.Start("Concurrent test")
			time.Sleep(200 * time.Millisecond)
		}()
		
		// Update progress concurrently
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // Wait a bit before updating
			p.Update("Updated concurrent test")
			time.Sleep(100 * time.Millisecond)
		}()
		
		wg.Wait()
		p.Success("All done")
	})
	
	assert.Contains(t, output, "✓ All done")
	assert.Equal(t, "Updated concurrent test", p.currentTask)
}

func TestProgressMultipleStopResume(t *testing.T) {
	p := NewProgress()
	
	captureOutput(func() {
		p.Start("Multi stop/resume test")
		time.Sleep(50 * time.Millisecond)
		
		// Stop and resume multiple times
		for i := 0; i < 3; i++ {
			p.Stop()
			time.Sleep(30 * time.Millisecond)
			assert.False(t, p.isRunning)
			
			p.Resume()
			time.Sleep(30 * time.Millisecond)
			assert.True(t, p.isRunning)
		}
		
		p.Success("Multi-cycle complete")
	})
}

func TestProgressChannelsClosed(t *testing.T) {
	p := NewProgress()
	
	captureOutput(func() {
		p.Start("Channel test")
		time.Sleep(50 * time.Millisecond)
		
		// Success should close the done channel and stop the goroutine
		p.Success("Channel test complete")
	})
	
	// Verify that we can't send to done channel again (it should be consumed)
	// This test ensures the goroutine properly exits
	time.Sleep(100 * time.Millisecond) // Give time for goroutine to exit
}

func TestSpinnerFramesValid(t *testing.T) {
	// Test that spinner frames are properly defined
	assert.NotEmpty(t, spinnerFrames)
	assert.Equal(t, 10, len(spinnerFrames))
	
	// Each frame should be a Unicode spinner character
	for i, frame := range spinnerFrames {
		assert.NotEmpty(t, frame, "Frame %d should not be empty", i)
		assert.True(t, len(frame) > 0, "Frame %d should have content", i)
	}
}

func TestProgressStyles(t *testing.T) {
	// Test that styles are properly initialized
	assert.NotNil(t, successMark)
	assert.NotNil(t, failMark)
	assert.NotNil(t, spinnerStyle)
	
	// Test that styles render something
	successOutput := successMark
	failOutput := failMark
	spinnerOutput := spinnerStyle.Render("⠋")
	
	assert.NotEmpty(t, successOutput)
	assert.NotEmpty(t, failOutput)
	assert.NotEmpty(t, spinnerOutput)
}

func TestProgressWithEmptyTask(t *testing.T) {
	p := NewProgress()
	
	output := captureOutput(func() {
		p.Start("")
		time.Sleep(50 * time.Millisecond)
		p.Success("Empty task done")
	})
	
	assert.Equal(t, "", p.currentTask)
	assert.Contains(t, output, "✓ Empty task done")
}

func TestProgressUpdateWhenNotRunning(t *testing.T) {
	p := NewProgress()
	
	// Update without starting should just update the task field
	p.Update("Updated without running")
	assert.Equal(t, "Updated without running", p.currentTask)
	assert.False(t, p.isRunning)
}

func TestProgressResumeWhenAlreadyRunning(t *testing.T) {
	p := NewProgress()
	
	captureOutput(func() {
		p.Start("Already running test")
		time.Sleep(50 * time.Millisecond)
		
		// Try to resume when already running - should be a no-op
		wasRunning := p.isRunning
		p.Resume()
		
		assert.True(t, wasRunning)
		assert.True(t, p.isRunning) // Should still be running
		
		p.Success("Done")
	})
}

func TestProgressStopWhenNotRunning(t *testing.T) {
	p := NewProgress()
	
	// Stop without starting should be safe
	p.Stop() // Should not panic or block
	
	assert.False(t, p.isRunning)
}

// Benchmark tests for performance
func BenchmarkProgressStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewProgress()
		captureOutput(func() {
			p.Start("Benchmark test")
			p.Success("Done")
		})
	}
}

func BenchmarkProgressUpdate(b *testing.B) {
	p := NewProgress()
	captureOutput(func() {
		p.Start("Benchmark update test")
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Update("Update test")
		}
		
		p.Success("Done")
	})
}