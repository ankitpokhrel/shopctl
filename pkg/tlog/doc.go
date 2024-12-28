// Package tlog provides a console-based structured logger for terminal applications.
//
// This package uses `log/slog`, available since go1.21+, as a logging frontend, and
// `uber-go/zap` as a logging backend. Technically, this package is simply a wrapper
// around these two libraries to display text based structured logs in the terminal.
//
// The package offers different levels of verbosity, allowing developers to control
// the amount of detail logged. The available verbosity levels (VerboseLevel) are:
//
//   - VL0: Default verbosity.
//   - VL1: Minimum verbosity.
//   - VL2: Intermediate verbosity.
//   - VL3: Highest verbosity.
//
// Example Usage:
//
//	logger := tlog.New(tlog.VL1) // VL1 is global VerboseLevel for this logger.
//
//	logger.Info("Application started") // VL0
//	logger.Infof("Application started with pid: %d", pid)
//	logger.V(tlog.VL2).Warn("Failed to execute query", "error", err) // Only displayed if global VerboseLevel is 2.
//
// The package uses colored level encoder by default.
package tlog
