package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestReady(t *testing.T) {
	r := &discordgo.Ready{
		Guilds: make([]*discordgo.Guild, 0),
	}

	Ready(dgTestBotSession, r)
}