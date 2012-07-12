package main

import (
	"database/sql"
	"github.com/bmizerany/pq"
	"log"
	"os"
)

func initdb() *sql.DB {
	connstr, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("pq.ParseURL", err)
	}
	db, err := sql.Open("postgres", connstr+" sslmode=disable")
	if err != nil {
		log.Fatal("sql.Open", err)
	}

	_, err = db.Exec(`
		create table if not exists release (
			plat text,
			cmd text,
			ver text,
			sha1 char(32) not null,
			primary key (plat, cmd, ver)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		create table if not exists cur (
			plat text,
			cmd text,
			curver text not null,
			foreign key (plat, cmd, curver) references release (plat, cmd, ver),
			primary key (plat, cmd)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		create table if not exists patch (
			plat text,
			cmd text,
			oldver text,
			newver text,
			sha1 char(32) not null,
			foreign key (plat, cmd, oldver) references release (plat, cmd, ver),
			foreign key (plat, cmd, newver) references release (plat, cmd, ver),
			primary key (plat, cmd, oldver, newver)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		create table if not exists next (
			plat text,
			cmd text,
			oldver text,
			newver text not null,
			foreign key (plat, cmd, oldver) references release (plat, cmd, ver),
			foreign key (plat, cmd, newver) references release (plat, cmd, ver),
			foreign key (plat, cmd, oldver, newver) references patch (plat, cmd, oldver, newver),
			primary key (plat, cmd, oldver)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
