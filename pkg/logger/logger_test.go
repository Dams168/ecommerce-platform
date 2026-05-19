package logger

import "testing"

func TestNew(t *testing.T){
    log := New("test-service")

    log.Info("This is an info message")
    log.Warn("This is a warning message")
    log.Error("This is an error message")

    t.Log("Logger Success")
}