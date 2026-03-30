package testutil

import (
	"io"
	"log"
	"testing"
)

// SilenceStandardLogger mutes standard logger output during tests/benchmarks.
func SilenceStandardLogger(tb testing.TB) {
	tb.Helper()

	oldWriter := log.Writer()
	oldFlags := log.Flags()
	oldPrefix := log.Prefix()

	log.SetOutput(io.Discard)
	log.SetFlags(0)
	log.SetPrefix("")

	tb.Cleanup(func() {
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
		log.SetPrefix(oldPrefix)
	})
}
