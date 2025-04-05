// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize sets up the logger with the specified debug level
func Initialize(debug bool) error {
	var cfg zap.Config
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	var err error
	log, err = cfg.Build()
	if err != nil {
		return err
	}

	return nil
}

// SetDebugMode dynamically changes the logger to debug mode
func SetDebugMode(debug bool) {
	// If logger is not initialized, initialize it
	if log == nil {
		if err := Initialize(debug); err != nil {
			// Since we can't log the error (logger isn't initialized),
			// we'll just print it to stderr
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		}
		return
	}
	
	// If we need to change to debug mode
	if debug {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		
		// Build a new logger
		newLog, err := cfg.Build()
		if err == nil {
			// Sync the old logger before replacing
			if err := log.Sync(); err != nil {
				// Ignore "inappropriate ioctl for device" errors which commonly happen
				// when stderr is a terminal device
				if !strings.Contains(err.Error(), "inappropriate ioctl for device") {
					fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
				}
			}
			log = newLog
		}
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if log != nil {
		log.Fatal(msg, fields...)
	}
}

// Sync flushes any buffered log entries
func Sync() error {
	if log != nil {
		err := log.Sync()
		// Ignore "inappropriate ioctl for device" errors which commonly happen
		// when stderr is a terminal device
		if err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			return err
		}
		return nil
	}
	return nil
} 