package dbManager

import (
	"github.com/jmoiron/sqlx"
	"fmt"
	"database/sql"
	"log"
	"net/url"
)

type DBInfo map[string]string

type DBManager struct {
	db *sqlx.DB
}

func NewDBManager(dbinfo map[string]string) (*DBManager, error) {
	str := fmt.Sprintf("%s://%s/%s?user=%s&password=%s&port=%s&sslmode=disable",
		dbinfo["engine"],
		dbinfo["host"],
		dbinfo["dbname"],
		dbinfo["username"],
		dbinfo["pass"],
		dbinfo["port"])

	db, err := sqlx.Open(dbinfo["engine"], str)

	if err != nil {
		return nil, err
	}

	return &DBManager{db: db}, nil
}

func (dbm *DBManager) CreateUser(values map[string]interface{}) (sql.Result, error) {

	result, err := dbm.db.Exec(`INSERT INTO users VALUES ($1, $2, $3) 
							ON CONFLICT ON CONSTRAINT table_name_pkey DO NOTHING;`,
		values["id"],
		values["age"],
		values["sex"])

	if err != nil {
		return nil, err
	}
	return result, nil

}

func (dbm *DBManager) GetStats(values url.Values) (*sql.Rows, error) {

	rows, err := dbm.db.Query(`SELECT
  date,
  id,
  age,
  cast(sex as VARCHAR(1)),
  cnt
FROM (
       SELECT
         *,
         row_number()
         OVER (
           PARTITION BY date
           ORDER BY cnt DESC) AS r
       FROM stats, users
       WHERE date >= $1
             AND date < $2
        AND "user" = id AND action = $3
     ) t
WHERE r <= $4
ORDER BY date, cnt DESC;`,
		values["date1"][0],
		values["date2"][0],
		values["action"][0],
		values["limit"][0])

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (dbm *DBManager) PutStats(values map[string]interface{}) (sql.Result, error) {

	result, err := dbm.db.Exec(`INSERT INTO stats ("user", action, date) VALUES ($1, $2, $3)
									  ON CONFLICT ON CONSTRAINT user_time_uniq
  									  DO UPDATE SET cnt = stats.cnt + 1;`,
		values["user"],
		values["action"],
		values["ts"])

	log.Print(err)

	if err != nil {
		return nil, err
	}
	return result, nil
}
