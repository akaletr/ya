package storage

import (
	"database/sql"
	"fmt"
	"log"
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
	rows, err := db.Query("select long from hello where short=? ", value)
	if err != nil {
		return "", err
	}

	// обязательно закрываем перед возвратом функции
	defer rows.Close()

	// пробегаем по всем записям
	l := ""
	for rows.Next() {
		rows.Scan(&l)
		fmt.Println(l)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return "", err
	}
	return l, nil
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

	_, err = db.Exec("create table data (id int, short varchar(30), long text)")
	if err != nil {
		return err
	}

	db.Exec("insert into data values (?, ?, ?)", id, key, value)
	return nil
}

func (p postgresDatabase) ReadAll(id string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func NewPostgresDatabase(connectionString string) Storage {
	return &postgresDatabase{
		connectionString: connectionString,
	}
}
