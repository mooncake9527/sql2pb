package cmd

import (
	"bytes"
	"fmt"
	tmpl "github.com/mooncake9527/sql2pb/template"
	"github.com/samber/lo"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type Converter struct {
	serverDbConfig       *DbConfig
	serverDb             *gorm.DB
	serverTable          *Table
	ignoreTable          *IgnoreTable
	serverTableColumns   []string
	serverTableColumnMap map[string]*MySQL2ProtoColumn
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
func NewConverter(serverDbConfig *DbConfig, serverDb *gorm.DB, serverTable *Table, ignoreTable *IgnoreTable) *Converter {
	return &Converter{
		serverDbConfig:       serverDbConfig,
		serverDb:             serverDb,
		serverTable:          serverTable,
		ignoreTable:          ignoreTable,
		serverTableColumnMap: make(map[string]*MySQL2ProtoColumn),
	}
}

// Start 启动。
func (c *Converter) Start() {
	defer wg.Done()
	switch c.serverTable.TableType {
	case "BASE TABLE":
		c.create()
	case "VIEW":
		slog.Warn(fmt.Sprintf("表 `%s` 不支持 VIEW 转换。", c.serverTable.TableName))
	}
}

// create 创建 PROTO。
func (c *Converter) create() {
	var (
		serverColumnData []Column
	)

	serverTableColumnResult := c.serverDb.Table("COLUMNS").Order("`ORDINAL_POSITION` ASC").Find(
		&serverColumnData, "`TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", c.serverDbConfig.Database, c.serverTable.TableName)

	if serverTableColumnResult.RowsAffected == 0 {
		return
	}

	primaryKey := ""
	PrimaryKeyProtoType := ""
	for _, serverColumn := range serverColumnData {
		_, find := lo.Find(c.ignoreTable.Columns, func(item string) bool {
			return item == serverColumn.ColumnName
		})
		if find {
			continue
		}
		dataType := strings.ToUpper(serverColumn.DataType)
		c.serverTableColumns = append(c.serverTableColumns, serverColumn.ColumnName)
		c.serverTableColumnMap[serverColumn.ColumnName] = &MySQL2ProtoColumn{
			DataType:      dataType,
			ProtoDataType: c.mappingDataTypeToProtoDataType(dataType),
		}
		if strings.Contains(serverColumn.ColumnKey, "PR") {
			primaryKey = serverColumn.ColumnName
			PrimaryKeyProtoType = serverColumn.DataType
		}
	}

	pt := ProtoTemplate{TableName: c.toCamelCase(c.serverTable.TableName)}
	pt.PrimaryKey = primaryKey
	pt.PrimaryKeyProtoType = c.mappingDataTypeToProtoDataType(PrimaryKeyProtoType)
	pt.PackageName = c.toCamelCaseFirstLower(c.serverDbConfig.Database)
	pt.ServiceName = c.toCamelCaseFirstLower(c.serverTable.TableName)
	for i, columnName := range c.serverTableColumns {
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

	if protoTpl != "" {
		tpl, err = template.ParseFiles(protoTpl)
		cobra.CheckErr(err)
	} else {
		tpl, err = template.New("").Parse(tmpl.ProtoTemplate)
		cobra.CheckErr(err)
	}
	var wr bytes.Buffer
	err = tpl.Execute(&wr, pt)
	cobra.CheckErr(err)
	_ = os.MkdirAll(out, os.ModePerm)
	err = os.WriteFile(fmt.Sprintf("%s/%s.proto", out, c.serverTable.TableName), wr.Bytes(), 0644)
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
