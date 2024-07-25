package cmd

import (
	"fmt"
	"github.com/spf13/cast"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Dsn         = "%s:%s@tcp(%s:%d)/information_schema?timeout=10s&parseTime=true&charset=%s"
	HostPattern = "^(.*)\\:(.*)\\@(.*)\\:(\\d+)$"
	DbPattern   = "^([A-Za-z0-9_]+)$"
	Charset     = "utf8mb4"
	//protoTemplate = "template/proto.tpl"
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVarP(&server, "server", "s", "", "指定服务器。(格式: <user>:<password>@<host>:<port>)")
	rootCmd.Flags().StringVarP(&db, "db", "d", "", "指定数据库。")
	rootCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "指定配置文件路径。")
	rootCmd.Flags().StringVarP(&out, "out", "o", "", "指定输出目录。")
	rootCmd.Flags().StringVarP(&tables, "tables", "t", "", "指定表名。")
	rootCmd.Flags().StringVarP(&protoTpl, "template", "p", "", "指定模板。")

	cobra.CheckErr(rootCmd.MarkFlagRequired("server"))
	cobra.CheckErr(rootCmd.MarkFlagRequired("db"))
	cobra.CheckErr(rootCmd.MarkFlagRequired("out"))
}

func initConfig() {
}

type Config struct {
	Ignores []*IgnoreTable `yaml:"ignores"`
}

type IgnoreTable struct {
	Table   string   `yaml:"table"`
	Columns []string `yaml:"columns"`
}

var (
	wg       sync.WaitGroup
	server   string
	db       string
	cfgPath  string
	out      string
	tables   string
	protoTpl string
)

var rootCmd = &cobra.Command{
	Use:     "proto",
	Short:   "mysql convert to proto.",
	Version: "v1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		serverMatched, err := regexp.MatchString(HostPattern, server)
		cobra.CheckErr(err)
		dbMatched, err := regexp.MatchString(DbPattern, db)
		cobra.CheckErr(err)
		if !serverMatched {
			cobra.CheckErr(fmt.Errorf("服务器 `%s` 格式错误。(正确格式: <user>:<password>@<host>:<port>)", server))
		}
		if !dbMatched {
			cobra.CheckErr(fmt.Errorf("数据库 `%s` 格式错误。", db))
		}

		var (
			serverUser = strings.Split(server[0:strings.LastIndex(server, "@")], ":")
			serverHost = strings.Split(server[strings.LastIndex(server, "@")+1:], ":")
		)
		serverDbConfig := &DbConfig{
			User:     serverUser[0],
			Password: serverUser[1],
			Host:     serverHost[0],
			Charset:  Charset,
			Database: db,
		}
		serverDbConfig.Port = cast.ToInt(serverHost[1])
		dsn := fmt.Sprintf(Dsn,
			serverDbConfig.User, serverDbConfig.Password,
			serverDbConfig.Host, serverDbConfig.Port,
			serverDbConfig.Charset,
		)
		serverDb, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
			SkipDefaultTransaction: true,
			DisableAutomaticPing:   true,
			Logger:                 logger.Default.LogMode(logger.Silent),
		})
		cobra.CheckErr(err)

		var serverSchema Schema
		serverSchemaResult := serverDb.Table("SCHEMATA").Limit(1).Find(
			&serverSchema,
			"SCHEMA_NAME = ?", serverDbConfig.Database,
		)
		if serverSchemaResult.RowsAffected <= 0 {
			cobra.CheckErr(fmt.Sprintf("数据库 `%s` 不存在。", serverDbConfig.Database))
		}

		var serverTableData []*Table
		queryServerTables := serverDb.Table("TABLES")
		if tables != "" {
			tableNames := strings.Split(tables, ",")
			queryServerTables.Where("TABLE_NAME in ?", tableNames)
		}
		serverTableResult := queryServerTables.Order("TABLE_NAME ASC").Find(&serverTableData, "TABLE_SCHEMA = ?", serverDbConfig.Database)
		if serverTableResult.RowsAffected <= 0 {
			cobra.CheckErr(fmt.Errorf("数据库 %s 没有表。", serverDbConfig.Database))
		}

		icMap := make(map[string]*IgnoreTable, 10)
		if cfgPath != "" {
			var ic *Config
			bytes, err := os.ReadFile(cfgPath)
			cobra.CheckErr(err)
			err = yaml.Unmarshal(bytes, &ic)
			cobra.CheckErr(err)

			for _, vv := range ic.Ignores {
				icMap[vv.Table] = vv
			}
		}

		for _, serverTable := range serverTableData {
			var ignoreTable = &IgnoreTable{}
			if v, ok := icMap[serverTable.TableName]; ok {
				ignoreTable = v
			}
			isContinue := true
			if ignoreTable.Table == serverTable.TableName && len(ignoreTable.Columns) == 0 {
				isContinue = false
			}
			if isContinue {
				wg.Add(1)
				go NewConverter(serverDbConfig, serverDb, serverTable, ignoreTable).Start()
			}
		}
		wg.Wait()
	},
}
