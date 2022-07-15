package storage

import (
	"fmt"
	"github.com/lib/pq"
	"log"

	"cmd/shortener/main.go/internal/model"

	sql "github.com/jmoiron/sqlx"
)

type postgresDatabase struct {
	db *sql.DB
}

func (p postgresDatabase) Read(value string) (string, error) {
	str := fmt.Sprintf("select long from data where short='%s'", value)
	rows, err := p.db.Query(str)
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
	str := fmt.Sprintf("insert into data values (%s, '%s', '%s')", id, key, value)
	_, err := p.db.Exec(str)

	if err != nil {
		// проверяем ошибку, если конфликт - возвращаем новую ошибку
		e, ok := err.(*pq.Error)
		if ok && e.Code == "23505" {
			return NewError(CONFLICT, e.Error())
		}
		return err
	}

	return nil
}

func (p postgresDatabase) WriteBatch(data model.DataBatch) error {
	_, err := p.db.NamedExec(`INSERT INTO data (id, short, long, correlation) 
		VALUES (:id, :short, :long, :correlation)`, data)

	if err != nil {
		// проверяем ошибку, если конфликт - возвращаем новую ошибку
		e, ok := err.(*pq.Error)
		if ok && e.Code == "23505" {
			return NewError(CONFLICT, e.Error())
		}
		return err
	}
	return nil
}

func (p postgresDatabase) ReadAll(id string) (map[string]string, error) {
	str := fmt.Sprintf("select short, long from data where id='%s'", id)
	rows, err := p.db.Query(str)
	if err != nil {
		log.Println(err)
		return map[string]string{}, err
	}

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

func (p postgresDatabase) Start() error {
	_, err := p.db.Exec("create table data (id varchar(30), short varchar(60) UNIQUE, long text, correlation varchar(30))")
	if err != nil {
		// проверяем ошибку, если ошибка "отношение "data" уже существует" все ок
		e, ok := err.(*pq.Error)
		if ok && e.Code == "42P07" {
			return nil
		}
		return err
	}

	return nil
}

func NewPostgresDatabase(connectionString string) (Storage, error) {
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		return &postgresDatabase{}, err
	}

	return &postgresDatabase{
		db: db,
	}, nil
}
