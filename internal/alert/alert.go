package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo    Level = "INFO"
	LevelWarning Level = "WARNING"
	LevelCritical Level = "CRITICAL"
)

// Alert holds information about a lease expiration event.
type Alert struct {
	LeaseID   string
	ExpiresIn time.Duration
	Level     Level
	Message   string
}

// Notifier is the interface implemented by alert senders.
type Notifier interface {
	Send(a Alert) error
}

// ConsoleNotifier writes alerts to an io.Writer (defaults to stdout).
type ConsoleNotifier struct {
	Out io.Writer
}

// NewConsoleNotifier returns a ConsoleNotifier writing to stdout.
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{Out: os.Stdout}
}

// Send formats and writes the alert to the configured writer.
func (c *ConsoleNotifier) Send(a Alert) error {
	_, err := fmt.Fprintf(
		c.Out,
		"[%s] lease=%s expires_in=%s message=%s\n",
		a.Level,
		a.LeaseID,
		a.ExpiresIn.Round(time.Second),
		a.Message,
	)
	return err
}

// Classify returns the alert Level based on expiry thresholds (in minutes).
func Classify(expiresIn time.Duration, warnMins, critMins int) Level {
	switch {
	case expiresIn <= time.Duration(critMins)*time.Minute:
		return LevelCritical
	case expiresIn <= time.Duration(warnMins)*time.Minute:
		return LevelWarning
	default:
		return LevelInfo
	}
}

// Build constructs an Alert for the given lease and expiry duration.
func Build(leaseID string, expiresIn time.Duration, warnMins, critMins int) Alert {
	lvl := Classify(expiresIn, warnMins, critMins)
	msg := fmt.Sprintf("Lease expires in %s", expiresIn.Round(time.Second))
	return Alert{
		LeaseID:   leaseID,
		ExpiresIn: expiresIn,
		Level:     lvl,
		Message:   msg,
	}
}
