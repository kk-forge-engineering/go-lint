// Package env reads the environment directly, which only infra may do.
package env

import (
	"fmt"
	"log" // want `log reads or writes log output past slog`
	"log/slog"
	_ "math/rand" // want `math/rand reads or writes random numbers`
	"os"
	clock "time"
)

// read reads what the program does not control. A method or a type of the same package is no read.
func read(logger *slog.Logger, at clock.Time) string {
	_ = clock.Since(at)               // want `time\.Since reads the clock`
	clock.Sleep(clock.Second)         // want `time\.Sleep reads the clock's timers`
	_ = os.Getenv("X")                // want `os\.Getenv reads the environment`
	_ = slog.New(slog.DiscardHandler) // want `log/slog\.New writes a logger of its own` `log/slog\.DiscardHandler writes a log handler`
	fmt.Println("x")                  // want `fmt\.Println writes standard output, which only infra and an entrypoint do`
	slog.Info("a feature may call slog's package-level functions")
	log.Print("x")
	logger.Info("x")
	_ = at.After(at)
	_ = clock.Duration(1)
	return fmt.Sprintf("%d", 1)
}
