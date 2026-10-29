// Package geo provides the 77 Thai provinces (kongvut/thai-province-data).
package geo

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// Province is a Thai province with English and Thai names.
type Province struct {
	NameEN string `json:"name_en"`
	NameTH string `json:"name_th"`
}

//go:embed provinces.json
var raw []byte

var (
	provinces []Province
	byLower   = map[string]string{}
)

func init() {
	if err := json.Unmarshal(raw, &provinces); err != nil {
		panic("geo: invalid provinces.json: " + err.Error())
	}
	for _, p := range provinces {
		byLower[strings.ToLower(p.NameEN)] = p.NameEN
		byLower[p.NameTH] = p.NameEN
	}
}

// List returns all provinces.
func List() []Province { return provinces }

// Canonical returns the canonical English province name for name (case-insensitive,
// Thai names accepted) and whether it is a known province.
func Canonical(name string) (string, bool) {
	n, ok := byLower[strings.ToLower(strings.TrimSpace(name))]
	return n, ok
}
