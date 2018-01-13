package atlas

import (
	"strings"

	"github.com/Cidan/gomud/interp"
	"github.com/Cidan/gomud/player"
	"github.com/rs/zerolog/log"
)

var PlayerLocation [][][][]map[string]*player.Player

func SetupPlayer() {
}

func StartPlayer(p *player.Player) {
	// Set the player up in the atlas and setup the interp
	// Why did I do this?!
	p.Interp = interp.NewLogin(p)

	for {
		str, err := p.Input.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("Error reading player input.")
			p.Stop()
			break
		}
		str = strings.TrimSpace(str)
		err = p.Interp.Read(str)
		if err != nil {
			log.Error().Err(err).
				Str("player", p.Data.UUID).
				Msg("Error reading input from player.")
		}
		log.Debug().Msg(str)
	}
}
