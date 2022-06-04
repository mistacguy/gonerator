package gonerator

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/mistacguy/gonerator/internal/db"
	"github.com/mistacguy/gonerator/internal/def"
	"github.com/mistacguy/gonerator/internal/model"
)

var wg sync.WaitGroup

type Gonerator struct {
	DB                *sql.DB
	url               string
	username          string
	password          string
	informationSchema string
	dbname            string
	tables            []string
	tableModels       model.Tables
	query             QueryStrategy
	packageName       string
}

func Mysql() *Gonerator {
	return &Gonerator{
		informationSchema: def.Schema,
	}
}

func (g *Gonerator) Username(username string) {
	g.username = username
}

func (g *Gonerator) Password(password string) {
	g.password = password
}

func (g *Gonerator) Connect(url string) {
	g.url = url
}

func (g *Gonerator) Database(dbname string) {
	g.dbname = dbname
}

func (g *Gonerator) Tables(tables []string) {
	g.tables = tables
}

func (g *Gonerator) Package(packageName string) {
	g.packageName = packageName
}

// QueryStrategy 数据库元信息查询接口
type QueryStrategy interface {
	tablesQuery()
}

// 查询数据库表元信息，返回指定数据库下指定的表的元信息
func queryTable(dbname string, tables []string, DB *sql.DB) model.Tables {
	tableModels := make(model.Tables, 0)
	queryStr := strings.Replace(def.DefaultQueryStr, "%db%", dbname, 1)
	for _, table := range tables {
		tableModel := model.Table{Name: table}
		tableQuery := strings.Replace(queryStr, "%table%", table, 1)
		rows, e := DB.Query(tableQuery)
		if e != nil {
			fmt.Println(e)
			return tableModels
		}
		columns := make(model.Columns, 0)
		for rows.Next() {
			column := model.Column{}
			e := rows.Scan(&column.Name, &column.DataType, &column.ColumnType, &column.Extra)
			if e == nil {
				columns = append(columns, column)
			}
		}
		tableModel.Columns = columns
		tableModels = append(tableModels, tableModel)
	}
	return tableModels
}

// DefaultQuery 默认查询策略
type DefaultQuery struct {
	g *Gonerator
}

// tablesQuery 实现默认查询策略的tableQuery方法
func (d *DefaultQuery) tablesQuery() {
	tableModels := queryTable(d.g.dbname, d.g.tables, d.g.DB)
	d.g.tableModels = tableModels

}

// NoTablesQuery 没有指定表名称的查询策略
type NoTablesQuery struct {
	g *Gonerator
}

// tablesQuery 实现没有指定表名称的查询策略策略的tablesQuery方法
func (n *NoTablesQuery) tablesQuery() {
	//获取指定数据库下的所有表
	queryStr := strings.Replace(def.TablesNameQueryStr, "%db%", n.g.dbname, 1)
	rows, e := n.g.DB.Query(queryStr)
	if e != nil {
		fmt.Println(e)
		return
	}
	tables := make([]string, 0)
	for rows.Next() {
		var table string
		e := rows.Scan(&table)
		if e == nil {
			tables = append(tables, table)
		}
	}
	tablesModel := queryTable(n.g.dbname, tables, n.g.DB)
	n.g.tables = tables
	n.g.tableModels = tablesModel
}

func (g *Gonerator) Generator() {
	// 拼接sql链接
	coonStr := g.username + ":" + g.password + "@tcp(" + g.url + ")/" + g.informationSchema + "?charset=utf8"
	if DB, DBConnectError := db.DB(coonStr); DBConnectError != nil {
		fmt.Println(DBConnectError)
		return
	} else {
		// 如果数据库连接成功，则获取数据库表信息
		if DBOpenError := DB.Ping(); DBOpenError != nil {
			fmt.Println(DBOpenError)
			return
		}
		g.DB = DB
		// 分配tableModels内存
		g.tableModels = make(model.Tables, 0)
		// 检查是否指定了目标数据库名称
		if g.dbname == "" {
			fmt.Println("没有指定数据库名称，请指定数据库名称")
			return
		}
		// 选择查询策略
		if len(g.tables) != 0 && g.tables != nil {
			g.query = &DefaultQuery{g: g}
		} else {
			g.query = &NoTablesQuery{g: g}
		}
		// 获取所有目标表的元信息
		g.query.tablesQuery()
		if g.packageName == "" {
			g.packageName = def.DefaultPackageName
		}

		templates := make(model.Templates, 0)
		for _, tableModel := range g.tableModels {
			template := table2Template(tableModel, g.packageName)
			templates = append(templates, template)
		}
		// 打开模板文件获取模板内容
		f := def.Template
		for _, template := range templates {
			wg.Add(1)
			go generating(template, f)
		}
		wg.Wait()
		//fmt.Println(AttrReplace(f, g.Tables[0], g.tableModels[g.Tables[0]]))
	}
}

func table2Template(table model.Table, packageName string) model.Template {
	template := model.Template{Package: packageName, Model: nameOpt(table.Name)}
	attributes := make([]string, 0)
	importsMap := make(map[string]string)
	for _, column := range table.Columns {
		goType := column2attribute(column.DataType)
		if goType == "time.Time" {
			importsMap["time.Time"] = "\"time\""
		}
		attribute := nameOpt(column.Name) + "\t" + column2attribute(column.DataType)

		attributes = append(attributes, attribute)
	}
	template.Attributes = attributes
	template.Imports = importsMap
	return template
}

func column2attribute(column string) string {
	var attr string
	switch column {
	case "int", "tinyint", "smallint", "mediumint", "bigint":
		attr = "int"
	case "varchar", "char", "text", "tinytext", "mediumtext", "longtext":
		attr = "string"
	case "datetime", "date", "time", "timestamp":
		attr = "time.Time"
	case "decimal", "float", "double", "real", "numeric":
		attr = "float64"
	case "bit", "bool":
		attr = "bool"
	case "json":
		attr = "string"
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		attr = "string"
	case "enum", "set":
		attr = "string"
	default:
		attr = "string"
	}
	return attr
}

func nameOpt(columnName string) string {
	name := columnName
	head := string(columnName[0])
	head = strings.ToUpper(head)
	name = head + columnName[1:]
	index := strings.Index(name, "_")
	for index != -1 {
		c := []rune(name)
		opt := strings.ToTitle(name[index+1 : index+2])
		o := []rune(opt)
		c[index+1] = o[0]
		name = string(c)
		name = strings.Replace(name, "_", "", 1)
		index = strings.Index(name, "_")
	}
	return name
}

//func (g *Gonerator) query(tableQueryMap map[string]string) {
//	for table, tableQuery := range tableQueryMap {
//		rows, e := g.DB.Query(tableQuery)
//		if e != nil {
//			fmt.Println(e)
//		}
//		tableModel := make([]model.Table, 0)
//		for rows.Next() {
//			table := model.Table{}
//			e := rows.Scan(&table.Tablename, &table.Columnname, &table.Datatype, &table.Columntype, &table.Extra)
//			if e == nil {
//				tableModel = append(tableModel, table)
//			}
//		}
//		g.tableModels[table] = tableModel
//	}
//}

// attr 获取属性字符串
func attr(template model.Template) string {
	var attrStr string
	for _, attr := range template.Attributes {
		attrStr += attr + note(attr) + "\n\t"

	}
	return attrStr
}

func imports(template model.Template) string {
	var importsStr string
	for _, importName := range template.Imports {
		importsStr += importName + "\n\t"
	}
	return importsStr
}

func note(attr string) string {
	temp := strings.Split(attr, "\t")[0]
	note := []rune(temp)
	note[0] = unicode.ToLower(note[0])
	result := "\t`json:\"" + string(note) + "\"`"
	return result
}

func fileName(modelName string) string {
	return strings.ToLower(modelName) + ".go"
}

// generating 根据模板生成文件
func generating(template model.Template, templateFile string) {
	defer wg.Done()
	templateFile = strings.Replace(templateFile, "%package_name%", template.Package, 1)
	templateFile = strings.Replace(templateFile, "%imports%", imports(template), 1)
	templateFile = strings.Replace(templateFile, "%model_name%", template.Model, 1)
	templateFile = strings.Replace(templateFile, "%attributes%", attr(template), 1)
	_, pathErr := os.Stat(template.Package)
	if pathErr != nil {
		if os.IsNotExist(pathErr) {
			os.Mkdir(template.Package, os.ModePerm)
		} else {
			fmt.Println(pathErr)
			return
		}
	}
	err := ioutil.WriteFile(template.Package+"/"+fileName(template.Model), []byte(templateFile), 0666)
	if err != nil {
		fmt.Println(err)
		fmt.Println("write fail: " + template.Model)
		return
	}
	fmt.Println("write success: " + template.Model)
}
