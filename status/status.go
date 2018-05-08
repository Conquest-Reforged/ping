package status

import (
	"encoding/json"
	"io"
	"net/http"
)

type Status struct {
	Description Description `json:"description"`
	Players     Players     `json:"players"`
	Version     Version     `json:"version"`
	ModInfo     *ModInfo    `json:"modinfo"`
}

type Description struct {
	Text string `json:"text"`
}

type Players struct {
	Max    int      `json:"max"`
	Online int      `json:"online"`
	Sample []Player `json:"sample"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type ModInfo struct {
	Type    string `json:"type"`
	ModList []Mod  `json:"modList"`
}

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Mod struct {
	ModID   string `json:"modid"`
	Version string `json:"version"`
}

func Get(url string) (*Status, error) {
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()
	return Decode(r.Body)
}

func Decode(r io.Reader) (*Status, error) {
	var status Status
	e := json.NewDecoder(r).Decode(&status)
	return &status, e
}

func (s *Status) Encode(w io.Writer) (error) {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	return e.Encode(s)
}