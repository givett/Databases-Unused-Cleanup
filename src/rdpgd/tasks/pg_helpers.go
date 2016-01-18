package tasks

import (
	_ "github.com/lib/pq"
	"github.com/starkandwayne/rdpgd/pg"
)

func getList(address string, sq string) (list []string, err error) {
	p := pg.NewPG(address, pbPort, `rdpg`, `rdpg`, pgPass)
	db, err := p.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows := []string{}
	err = db.Select(&rows, sq)
	if err != nil {
		return nil, err
	}
	return rows, nil

}

func execQuery(address string, sq string) (err error) {
	p := pg.NewPG(address, pbPort, `rdpg`, `rdpg`, pgPass)
	db, err := p.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(sq)
	return err
}
