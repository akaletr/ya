package storage

import (
	"cmd/shortener/main.go/internal/model"
	//"database/sql"
	"fmt"
	"log"

	sql "github.com/jmoiron/sqlx"
)

type postgresDatabase struct {
	connectionString string
}

func (p postgresDatabase) Read(value string) (string, error) {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return "", err
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	str := fmt.Sprintf("select long from data where short='%s'", value)
	rows, err := db.Query(str)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	// пробегаем по всем записям
	long := ""
	for rows.Next() {
		err = rows.Scan(&long)
		if err != nil {
			return "", err
		}
	}

	err = rows.Err()
	if err != nil {
		return "", err
	}

	return long, nil
}

func (p postgresDatabase) Write(id, key, value string) error {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return err
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	// если таблицы нет - создаем
	_, err = db.Exec("create table data (id varchar(30), short varchar(60), long text, correlation varchar(30))")
	if err != nil {
		fmt.Println(err)
	}
	str := fmt.Sprintf("insert into data values (%s, '%s', '%s')", id, key, value)
	_, err = db.Exec(str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p postgresDatabase) WriteBatch(data model.DataBatch) error {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return err
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	_, err = db.NamedExec(`INSERT INTO data (id, short, long, correlation) 
		VALUES (:id, :short, :long, :correlation)`, data)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func (p postgresDatabase) ReadAll(id string) (map[string]string, error) {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return map[string]string{}, err
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	str := fmt.Sprintf("select short, long from data where id='%s'", id)
	rows, err := db.Query(str)
	if err != nil {
		log.Println(err)
		return map[string]string{}, err
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	result := map[string]string{}
	for rows.Next() {
		var short, long string
		err = rows.Scan(&short, &long)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result[short] = long
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func NewPostgresDatabase(connectionString string) Storage {
	return &postgresDatabase{
		connectionString: connectionString,
	}
}
