package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func DbConnection() *xorm.Engine {
	var db *xorm.Engine
	var err error

	for {

		db, err = xorm.NewEngine("sqlite3", "generator.db")
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}

		if err := db.Ping(); err != nil {
			log.Println("--",err)
				time.Sleep(time.Second)
			continue
		}

		break
	}

	if err := migration(db); err != nil {
		log.Fatalln(err)
	}
	return db
}

func migration(db *xorm.Engine) error {
	driver, err := sqlite3.WithInstance(db.DB().DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failure migration driver %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance("file://migrations", "generator.db", driver)
		log.Println(err)
	if err != nil {
		return fmt.Errorf("failure migration file %w", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failure migration up %w", err)
	}
	return nil
}
