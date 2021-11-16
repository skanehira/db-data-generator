package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/rs/xid"
	flag "github.com/spf13/pflag"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

type Column struct {
	Name  string      `yaml:"name"`
	Type  string      `yaml:"type"`
	Value interface{} `yaml:"value"`
}

type Table struct {
	Name    string   `yaml:"name"`
	Columns []Column `yaml:"columns"`
}

type Database struct {
	Database string  `yaml:"database"`
	Tables   []Table `yaml:"tables"`
}

var rootCmd = &cobra.Command{
	Use: "db-data-generator",
}

func exitError(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func mustGetFlag(flag *flag.FlagSet, name string) string {
	v, err := flag.GetString(name)
	if err != nil {
		exitError(err)
	}
	return v
}

func insert(tx *sql.Tx, db Database, limit int) error {
	rand.Seed(time.Now().UnixNano())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, t := range db.Tables {
		cols := make([]string, len(t.Columns))
		for i, cn := range t.Columns {
			cols[i] = cn.Name
		}

		var start int
		offset := 10000

		for start < limit {
			q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES `, t.Name, strings.Join(cols, ","))

			vs := make([]string, offset)
			for i := 0; i < offset; i++ {
				var values []string

				for _, col := range t.Columns {
					switch v := col.Value.(type) {
					case string:
						if v == "xid" {
							values = append(values, `"`+xid.New().String()+`"`)
						} else {
							values = append(values, v)
						}
					case []interface{}:
						switch vv := v[rand.Intn(len(v))].(type) {
						case string:
							values = append(values, `"`+vv+`"`)
						case int:
							values = append(values, strconv.Itoa(vv))
						}
					}
				}
				vs[i] = "(" + strings.Join(values, ",") + ")"
			}
			q += strings.Join(vs, ",")

			if _, err := tx.ExecContext(ctx, q); err != nil {
				return err
			}

			start += offset
		}
	}
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	flag := cmd.PersistentFlags()
	host := mustGetFlag(flag, "host")
	if host == "" {
		host = "localhost"
	}
	user := mustGetFlag(flag, "user")
	pass := mustGetFlag(flag, "password")
	port := mustGetFlag(flag, "port")
	if port == "" {
		port = "3306"
	}
	dbname := mustGetFlag(flag, "database")
	file := mustGetFlag(flag, "file")
	limitp := mustGetFlag(flag, "limit")
	limit, err := strconv.Atoi(limitp)
	if err != nil {
		return err
	}

	var setting Database

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&setting); err != nil {
		return err
	}

	db, err := sql.Open(setting.Database, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=UTC",
		user, pass, host, port, dbname))
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	log.Println("start insert...")
	defer log.Println("finished insert")
	if err := insert(tx, setting, limit); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func Execute() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	rootCmd.PersistentFlags().StringP("host", "", "", "database host")
	rootCmd.PersistentFlags().StringP("user", "u", "", "user name")
	rootCmd.PersistentFlags().StringP("password", "p", "", "password")
	rootCmd.PersistentFlags().StringP("port", "P", "", "database port")
	rootCmd.PersistentFlags().StringP("database", "D", "", "database name")
	rootCmd.PersistentFlags().StringP("file", "f", "", "setting file")
	rootCmd.PersistentFlags().StringP("limit", "l", "", "gererate table rows limit")

	if err := rootCmd.Execute(); err != nil {
		exitError(err)
	}
}
