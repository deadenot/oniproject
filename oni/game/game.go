package game

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/sessions"
	"net/http"
	"oniproject/oni/utils"
	"time"
)

type AvatarDB interface {
	AvatarById(utils.Id) (*Avatar, error)
	SaveAvatar(*Avatar) error
}

type GameAddr interface {
	GameAddr() string
}

type Game struct {
	Map *Map
	adb AvatarDB
}

func NewGame(adb AvatarDB) *Game {
	game := &Game{adb: adb}
	game.Map = NewMap(game)
	return game
}

func (gm *Game) Run(addr string) {
	log.Println("run GAME:", addr)

	// TODO: init RPC

	go gm.Map.Run()
	// run http server

	m := martini.Classic()
	m.Map(utils.CreateMartiniLogger())

	store := sessions.NewCookieStore([]byte("secret123"))
	store.Options(sessions.Options{Path: "/", MaxAge: 86400 * 30, HttpOnly: true})
	m.Use(sessions.Sessions("ssid", store))

	m.Get("/ws", func(sessions sessions.Session, w http.ResponseWriter, r *http.Request) (int, string) {
		_id := sessions.Get("id")
		if _id == nil {
			log.Warn("Unauthorized ", sessions)
			return 401, "Unauthorized"
		}
		id := _id.(int64)
		log.Debug("_ID ", _id, " id ", id)

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			if _, ok := err.(websocket.HandshakeError); !ok {
				return 500, "fail HandshakeError"
			}
			return 418, http.StatusText(418)
		}

		a, err := gm.adb.AvatarById(utils.Id(id))
		if err != nil {
			log.Error("get avatar", err)
			return 500, http.StatusText(418)
		}

		a.Connection = Connection{
			ws:          ws,
			sendMessage: make(chan interface{}, 2560),
			ping_pong:   time.Now(),
		}
		log.Debug(a)

		gm.Map.RunAvatar(ws, a)

		return 200, "game over"
	})

	http.Handle("/ws", m)

	if addr != "" {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}

func (gm *Game) LoadMap(id utils.Id) {
	log.Println("LoadMap", id)
}
func (gm *Game) UnloadMap(id utils.Id) {
	log.Println("UnloadMap", id)
}

func (gm *Game) DetachAvatar(id utils.Id, mapId int64) error {
	obj := gm.Map.GetObjById(id)
	if obj == nil {
		return errors.New("Avatar not found")
	}
	if avatar, ok := obj.(*Avatar); ok {
		gm.Map.Unregister(avatar)
		return nil
	}
	return errors.New("Avatar not found")
}