package oni

// todo mutex for avatars

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gocircuit/circuit/client"
	"oniproject/oni/game"
	"oniproject/oni/utils"
)

type Balancer struct {
	c     *client.Client
	games []*BalancerGame
	adb   AvatarDB
}

type BalancerMap struct {
	Max     int
	Avatars map[utils.Id]bool
}

type BalancerGame struct {
	Addr         string
	Rpc          string
	Minid, Maxid int64
	Maps         map[string]*BalancerMap
	game         *game.Game
}

func (g *BalancerGame) LoadMap(id string) {
	g.game.LoadMap(id)
}

func NewBalancer(config *Config, adb AvatarDB) (b *Balancer) {
	b = &Balancer{
		adb:   adb,
		games: config.Games,
	}
	if config.Circuit != "" {
		b.c = client.Dial(config.Circuit, nil)
	}
	return
}

func (b *Balancer) AttachAvatar(id utils.Id) (host string, mapId string, a *game.Avatar, err error) {
	a, err = b.adb.AvatarById(id)
	if err != nil {
		a = nil
		return
	}

	m, game := b.findMap(a.MapId)

	if _, ok := m.Avatars[a.Id()]; ok {
		game.game.DetachAvatar(a.Id(), a.MapId)
	}

	m.Avatars[a.Id()] = true
	host = game.Addr
	mapId = a.MapId

	// attach

	return
}

/* TODO
func (b *Balancer) DetachAvatar(a *game.Avatar) error {
	if m, ok := b.Maps[utils.Id(a.MapId)]; ok {
		if _, ok := m.Avatars[a.Id()]; ok {
			delete(m.Avatars, a.Id())
			// TODO send it to Game
		}
	}
	return nil
}*/

func (b *Balancer) findMap(id string) (*BalancerMap, *BalancerGame) {
	for _, g := range b.games {
		if m, ok := g.Maps[id]; ok {
			if len(m.Avatars) >= m.Max {
				log.Errorf("Map is full %v %v", m, g)
				continue
			}
			return m, g
		}
	}

	log.Info("create map ", id)

	// FIXME find free Map

	g := b.games[0]
	if g.game == nil {
		// XXX
		g.game = game.NewGame(b.adb)
		go g.game.Run("")
	}
	if g.Maps == nil {
		g.Maps = make(map[string]*BalancerMap)
	}

	m := &BalancerMap{
		Max:     2000,
		Avatars: make(map[utils.Id]bool),
	}
	g.Maps[id] = m
	g.LoadMap(id)

	return m, g
}
