package api_test

import "github/M2A96/Monopoly.git/log"

// noopLogger is a do-nothing RuntimeLogger for use in unit tests.
type noopLogger struct{}

func (n *noopLogger) Debug(_ string, _ ...any)                      {}
func (n *noopLogger) Info(_ string, _ ...any)                       {}
func (n *noopLogger) Warn(_ string, _ ...any)                       {}
func (n *noopLogger) Error(_ string, _ ...any)                      {}
func (n *noopLogger) Fatal(_ string, _ ...any)                      {}
func (n *noopLogger) WithField(_ string, _ any) log.RuntimeLogger   { return n }
func (n *noopLogger) WithFields(_ map[string]any) log.RuntimeLogger { return n }
func (n *noopLogger) Fields() map[string]any                        { return nil }
