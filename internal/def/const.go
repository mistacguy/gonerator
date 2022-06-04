package def

const (
	Schema             = "information_schema"
	DefaultQueryStr    = "select COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, EXTRA from COLUMNS where TABLE_SCHEMA = '%db%' and TABLE_NAME = '%table%'"
	TablesNameQueryStr = "select TABLE_NAME from TABLES where TABLE_SCHEMA = '%db%'"
)

const (
	DefaultPackageName = "model"
)

const Template = `
package %package_name%

import (
    %imports%
)

type %model_name% struct {
    %attributes%
}
`
