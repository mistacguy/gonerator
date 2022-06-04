package model

type Template struct {
	Package    string   `json:"package"`
	Model      string   `json:"model"`
	Attributes []string `json:"attributes"`
	Imports    Imports  `json:"imports"`
}

type Templates []Template
type Imports map[string]string
