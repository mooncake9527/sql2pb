package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/mooncake9527/sql2pb/config"
	"github.com/mooncake9527/x/xerrors/xerror"

	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Dsn = "%s:%s@tcp(%s:%d)/information_schema?timeout=10s&parseTime=true&charset=%s"
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "指定配置文件路径。")
	cobra.CheckErr(rootCmd.MarkFlagRequired("config"))
}

type Config struct {
	Ignores []*IgnoreTable `yaml:"ignores"`
}

type IgnoreTable struct {
	Table   string   `yaml:"table"`
	Columns []string `yaml:"columns"`
}

var (
	wg      sync.WaitGroup
	cfgPath string
)

var rootCmd = &cobra.Command{
	Use:     "proto",
	Short:   "mysql convert to proto.",
	Version: "v1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
		}
	},
}

func run() error {
	config.Parse(cfgPath)
	dsn := fmt.Sprintf(Dsn,
		config.AppConfig.DB.User,
		config.AppConfig.DB.Password,
		config.AppConfig.DB.Host,
		config.AppConfig.DB.Port,
		"utf8mb4",
	)
	informationSchemaConn, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return xerror.New(err.Error())
	}

	var schema Schema
	r := informationSchemaConn.Table("SCHEMATA").Limit(1).Find(&schema, "SCHEMA_NAME = ?", config.AppConfig.DB.Schema)
	if r.RowsAffected <= 0 {
		return xerror.Newf("数据库 `%s` 不存在。", config.AppConfig.DB.Schema)
	}

	var tables []*Table
	q := informationSchemaConn.Table("TABLES")
	if config.AppConfig.DB.Tables != "" {
		tableNames := strings.Split(config.AppConfig.DB.Tables, ",")
		q.Where("TABLE_NAME in ?", tableNames)
	}
	r = q.Order("TABLE_NAME ASC").Find(&tables, "TABLE_SCHEMA = ?", config.AppConfig.DB.Schema)
	if r.RowsAffected <= 0 {
		return xerror.Newf("数据库 `%s` 没有表。", config.AppConfig.DB.Schema)
	}

	_ = os.MkdirAll(config.AppConfig.Out, os.ModePerm)

	for _, table := range tables {
		wg.Add(1)
		go NewConverter(informationSchemaConn, table).Start()
	}
	wg.Wait()

	return nil
}
