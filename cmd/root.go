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

type Definition struct {
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

func parseValue(i interface{}) (string, error) {
	switch v := i.(type) {
	case string:
		if v == "xid" {
			return `"` + xid.New().String() + `"`, nil
		} else {
			return `"` + v + `"`, nil
		}
	case uint64, uint32, uint8, uint, int, int8, int32, int64:
		return fmt.Sprintf("%v", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case []interface{}:
		return parseValue(v[rand.Intn(len(v))])
	}
	return "", fmt.Errorf("invalid value: %v", i)
}

func insert(db *sql.DB, def Definition, limit int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, t := range def.Tables {
		cols := make([]string, len(t.Columns))
		for i, cn := range t.Columns {
			cols[i] = cn.Name
		}
		q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES `, t.Name, strings.Join(cols, ","))
		var vs []string

		for i := 0; i < limit; i++ {
			if i > 0 && i%1000 == 0 {
				if _, err := tx.ExecContext(ctx, q+strings.Join(vs, ",")); err != nil {
					return err
				}
				vs = []string{}
			}
			values := make([]string, len(t.Columns))
			for j, col := range t.Columns {
				v, err := parseValue(col.Value)
				if err != nil {
					return err
				}
				values[j] = v
			}
			vs = append(vs, "("+strings.Join(values, ",")+")")
		}

		if _, err := tx.ExecContext(ctx, q+strings.Join(vs, ",")); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
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

	var def Definition

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&def); err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=UTC",
		user, pass, host, port, dbname)
	db, err := sql.Open(def.Database, dsn)

	if err != nil {
		return err
	}
	defer db.Close()

	log.Println("start insert...")
	defer log.Println("finished insert")
	if err := insert(db, def, limit); err != nil {
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
	rootCmd.PersistentFlags().StringP("file", "f", "", "definition file")
	rootCmd.PersistentFlags().StringP("limit", "l", "", "generate table rows limit")

	if err := rootCmd.Execute(); err != nil {
		exitError(err)
	}
}
