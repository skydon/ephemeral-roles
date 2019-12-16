package monitor

import (
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/discordbotsorg"
	"github.com/prometheus/client_golang/prometheus"
)

type guilds struct {
	Log                 *logrus.Logger
	Session             *discordgo.Session
	HTTPClient          *http.Client
	BotID               string
	DiscordBotsOrgToken string
	PrometheusGauge     prometheus.Gauge
	Interval            time.Duration
	cache               *guildsCache
}

type guildsCache struct {
	mutex     sync.Mutex
	guildList []*discordgo.Guild
	numGuilds int
}

// Monitor sets up an infinite loop checking guild changes
func (g *guilds) Monitor() {
	g.cache = &guildsCache{}

	for {
		g.update()
		time.Sleep(g.Interval)
	}
}

func (g *guilds) update() {
	g.cache.mutex.Lock()
	defer g.cache.mutex.Unlock()

	botName := g.Session.State.User.Username

	originalCount := g.cache.numGuilds
	newCount := len(g.Session.State.Guilds)

	switch {
	case newCount == originalCount:
		return
	case newCount > originalCount:
		newGuild := g.Session.State.Guilds[newCount-1]
		g.Log.WithField("guild", newGuild.Name).Info(botName + " joined new guild")
	case newCount < originalCount:
		g.Log.Info(botName + " removed from guild")
	}

	g.cache.numGuilds = newCount
	g.cache.guildList = g.Session.State.Guilds
	g.PrometheusGauge.Set(float64(newCount))

	// discordbots.org integration
	if g.BotID != "" && g.DiscordBotsOrgToken != "" {
		err := discordbotsorg.Update(
			g.HTTPClient,
			g.DiscordBotsOrgToken,
			g.BotID,
			newCount,
		)
		if err != nil {
			g.Log.WithError(err).Warnf("unable to update guild count")
			return
		}
	}
}
