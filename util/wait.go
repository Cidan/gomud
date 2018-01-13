package util

import (
	"os"
	"os/signal"
)

// WaitForSigInt waits for... sigint
func SigIntChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	return c
}
