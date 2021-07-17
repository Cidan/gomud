package construct

import (
	"bufio"
	"context"
	"testing"

	"github.com/Cidan/gomud/lock"
	"github.com/stretchr/testify/assert"
)

// This test rapidly moves a player between two rooms, while deleting a room in
// another go routine, ensuring room deletes are concurrently safe.
// This test should be run with race detection via `go test -race -count=3`
func TestPlayerMovementRace(t *testing.T) {
	testSetupWorld(t)
	r, w := testLoginNewUser(t, "Playerm")

	c := make(chan bool)
	go func(reader *bufio.Reader, writer *bufio.Writer) {
		w.WriteString("build\n")
		assert.Nil(t, writer.Flush())
		for i := 0; i < 100; i++ {
			writer.WriteString("dig east\n")
			assert.Nil(t, writer.Flush())
			writer.WriteString("west\n")
			assert.Nil(t, writer.Flush())
		}
		c <- true
	}(r, w)

	ctx := lock.Context(context.Background(), "room_deleter")
	for i := 0; i < 10000; i++ {
		r := Atlas.GetRoom(1, 0, 0)
		if r != nil {
			r.Delete(ctx)
		}
	}
	<-c
}

func TestEditRoom(t *testing.T) {
	testSetupWorld(t)
	_, w := testLoginNewUser(t, "EditRoom")
	runCommands(t, nil, w, []string{
		"build",
		"edit room description",
		":w",
	})
}
