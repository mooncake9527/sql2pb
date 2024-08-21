package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/mooncake9527/sql2pb/config"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type Converter struct {
	informationSchemaConn *gorm.DB
	table                 *Table
	tableColumns          []string
	serverTableColumnMap  map[string]*MySQL2ProtoColumn
}

type MySQL2ProtoColumn struct {
	DataType      string
	ProtoDataType string
}

type ProtoTemplate struct {
	PackageName         string // eg. companyUser
	ServiceName         string // eg. companyUser
	TableName           string // eg. companyUser
	PrimaryKey          string // eg. id
	PrimaryKeyProtoType string // eg. int64
	ProtoColumns        []ProtoColumn
}

type ProtoColumn struct {
	ColumnName string
	FieldName  string
	ColumnType string
	ColumnNum  int32
}

// NewConverter 新建转换器。
func NewConverter(serverDb *gorm.DB, serverTable *Table) *Converter {
	return &Converter{
		informationSchemaConn: serverDb,
		table:                 serverTable,
		serverTableColumnMap:  make(map[string]*MySQL2ProtoColumn),
	}
}

func (c *Converter) Start() {
	defer wg.Done()
	switch c.table.TableType {
	case "BASE TABLE":
		c.create()
	case "VIEW":
		slog.Warn(fmt.Sprintf("表 `%s` 不支持 VIEW 转换。", c.table.TableName))
	}
}

func (c *Converter) create() {
	var (
		columns []Column
	)

	r := c.informationSchemaConn.Table("COLUMNS").Order("`ORDINAL_POSITION` ASC").Find(
		&columns, "`TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", config.AppConfig.DB.Schema, c.table.TableName)

	if r.RowsAffected == 0 {
		return
	}

	primaryKey := ""
	PrimaryKeyProtoType := ""
	for _, column := range columns {
		dataType := strings.ToUpper(column.DataType)
		c.tableColumns = append(c.tableColumns, column.ColumnName)
		c.serverTableColumnMap[column.ColumnName] = &MySQL2ProtoColumn{
			DataType:      dataType,
			ProtoDataType: c.mappingDataTypeToProtoDataType(dataType),
		}
		if strings.Contains(column.ColumnKey, "PR") {
			primaryKey = column.ColumnName
			PrimaryKeyProtoType = column.DataType
		}
	}

	pt := ProtoTemplate{TableName: c.toCamelCase(c.table.TableName)}
	pt.PrimaryKey = primaryKey
	pt.PrimaryKeyProtoType = c.mappingDataTypeToProtoDataType(PrimaryKeyProtoType)
	pt.PackageName = c.toCamelCaseFirstLower(config.AppConfig.DB.Schema)
	pt.ServiceName = c.toCamelCaseFirstLower(c.table.TableName)
	for i, columnName := range c.tableColumns {
		if pc, ok := c.serverTableColumnMap[columnName]; ok {
			protoColumn := ProtoColumn{
				ColumnName: columnName,
				FieldName:  c.toCamelCaseFirstLower(columnName),
				ColumnType: pc.ProtoDataType,
				ColumnNum:  int32(i + 1),
			}
			pt.ProtoColumns = append(pt.ProtoColumns, protoColumn)
		}
	}

	var (
		tpl *template.Template
		err error
	)

	f,err:=os.Open(config.AppConfig.Tpl)
	cobra.CheckErr(err)
	b,err:=io.ReadAll(f)
	cobra.CheckErr(err)
	tpl, err = template.New(config.AppConfig.Tpl).Parse(string(b))
	cobra.CheckErr(err)
	var wr bytes.Buffer
	err = tpl.Execute(&wr, pt)
	cobra.CheckErr(err)
	_ = os.MkdirAll(config.AppConfig.Out, os.ModePerm)
	err = os.WriteFile(fmt.Sprintf("%s/%s.proto", config.AppConfig.Out, c.table.TableName), wr.Bytes(), 0644)
	cobra.CheckErr(err)
}

func (c *Converter) mappingDataTypeToProtoDataType(dataType string) string {
	dataType = strings.ToUpper(dataType)
	switch dataType {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER":
		return "int32"
	case "BIGINT":
		return "int64"
	case "FLOAT", "DECIMAL":
		return "float32"
	case "DOUBLE":
		return "float64"
	case "DATE", "TIME", "YEAR", "DATETIME", "TIMESTAMP", "CHAR", "VARCHAR", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT":
		return "string"
	case "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB":
		return "bytes"
	}
	return "string"
}

func (c *Converter) toCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i := 0; i < len(words); i++ {
		words[i] = cases.Title(language.English).String(words[i])
	}
	return strings.Join(words, "")
}

func (c *Converter) toCamelCaseFirstLower(s string) string {
	words := strings.Split(s, "_")
	for i := 0; i < len(words); i++ {
		if i == 0 {
			words[i] = cases.Lower(language.English).String(words[i])
		} else {
			words[i] = cases.Title(language.English).String(words[i])
		}
	}
	return strings.Join(words, "")
}
