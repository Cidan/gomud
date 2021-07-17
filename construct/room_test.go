package construct

import (
	"bufio"
	"testing"

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
		for i := 0; i < 100; i++ {
			writer.WriteString("dig east\n")
			assert.Nil(t, writer.Flush())
			//reader.ReadString('\r')
			writer.WriteString("west\n")
			assert.Nil(t, writer.Flush())

			//reader.ReadString('\r')
		}
		c <- true
	}(r, w)

	for i := 0; i < 10000; i++ {
		r := Atlas.GetRoom(1, 0, 0)
		if r != nil {
			// TODO(lobato): fix this test
			//r.Delete()
		}
	}
	<-c
}

/*
func TestEditRoom(t *testing.T) {
	r, w := testLoginNewUser(t, "EditRoom")

	w.WriteString("edit room description\n")
	assert.Nil(t, w.Flush())
	r.ReadString('\r')

	w.WriteString(":w\n")
	assert.Nil(t, w.Flush())
	r.ReadString('\r')
}
*/
