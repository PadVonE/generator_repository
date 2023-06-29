package main

import (
	"fmt"
	"github.com/2q4t-plutus/envopt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	gp "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func DbConnection() *gorm.DB {
	var db *gorm.DB
	var err error

	for {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			envopt.GetEnv("POSTGRES_HOST"),
			envopt.GetEnv("POSTGRES_PORT"),
			envopt.GetEnv("POSTGRES_USER"),
			envopt.GetEnv("POSTGRES_PASSWORD"),
			envopt.GetEnv("POSTGRES_DB_NAME"),
			envopt.GetEnv("POSTGRES_SSL_MODE"),
			envopt.GetEnv("TZ"),
		)

		config := &gorm.Config{
			//Logger: logger.Default.LogMode(logger.Info),
		}

		db, err = gorm.Open(gp.Open(dsn), config)
		if err != nil {
			log.Warn(err)
			time.Sleep(time.Second)
			continue
		}

		break
	}

	if err := migration(db); err != nil {
		log.Warn(err)
	}

	return db
}

func migration(db *gorm.DB) error {
	sql, err := db.DB()
	if err != nil {
		return fmt.Errorf("failure migration sql %w", err)
	}

	driver, err := postgres.WithInstance(sql, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failure migration driver %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance("file://migrations", envopt.GetEnv("POSTGRES_DB_NAME"), driver)
	if err != nil {
		return fmt.Errorf("failure migration file %w", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failure migration up %w", err)
	}
	return nil
}
