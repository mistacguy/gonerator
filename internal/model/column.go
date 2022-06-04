package model

type Column struct {
	Name       string `json:"name"`
	DataType   string `json:"datatype"`
	ColumnType string `json:"columntype"`
	Extra      string `json:"extra"`
}

type Columns []Column
