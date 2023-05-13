package mpris

import (
	"sptlrx/player"
	"strings"

	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
)

func New(name string) (*Client, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	var names []string
	if name != "" {
		names = strings.Split(name, ",")
	}
	return &Client{names, conn}, nil
}

// Client implements player.Player
type Client struct {
	names []string
	conn  *dbus.Conn
}

func (p *Client) getPlayer() (*mpris.Player, error) {
	players, err := mpris.List(p.conn)
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, nil
	}

	if len(p.names) == 0 {
		return mpris.New(p.conn, players[0]), nil
	}

	// iterating over configured names
	for _, name := range p.names {
		for _, player := range players {
			// trim "org.mpris.MediaPlayer2."
			if player[23:] == name {
				return mpris.New(p.conn, player), nil
			}
		}
	}
	return nil, nil
}

func (p *Client) State() (*player.State, error) {
	pl, err := p.getPlayer()
	if err != nil {
		return nil, err
	}
	if pl == nil {
		return nil, nil
	}

	status, err := pl.GetPlaybackStatus()
	if err != nil {
		return nil, err
	}
	position, err := pl.GetPosition()
	if err != nil {
		// unsupported player
		return nil, err
	}
	meta, err := pl.GetMetadata()
	if err != nil {
		return nil, err
	}

	var title string
	if t, ok := meta["xesam:title"].Value().(string); ok {
		title = t
	}

	var artist string
	switch a := meta["xesam:artist"].Value(); a.(type) {
	case string:
		artist = a.(string)
	case []string:
		artist = strings.Join(a.([]string), " ")
	}

	var query string
	if artist != "" {
		query = artist + " " + title
	} else {
		query = title
	}

	return &player.State{
		ID:       query, // use query as id since mpris:trackid is broken
		Query:    query,
		Position: int(position * 1000), // secs to ms
		Playing:  status == mpris.PlaybackPlaying,
	}, err
}
