package storage

import (
	sql1 "database/sql"
	"log"

	"cmd/shortener/main.go/internal/model"

	sql "github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresDatabase struct {
	db *sql.DB
}

func (p postgresDatabase) Read(value string) (model.Note, error) {
	rows, err := p.db.Query("select long, deleted from data where short=$1", value)
	if err != nil {
		log.Println(err)
		return model.Note{}, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	// пробегаем по всем записям
	long := ""
	// беру из пакета database/sql тип для чтения типа bool из базы
	deleted := sql1.NullBool{}

	for rows.Next() {
		err = rows.Scan(&long, &deleted)
		if err != nil {
			return model.Note{}, err
		}
	}

	err = rows.Err()
	if err != nil {
		return model.Note{}, err
	}

	return model.Note{
		Long:    long,
		Deleted: deleted.Bool,
	}, nil
}

func (p postgresDatabase) Write(note model.Note) error {
	_, err := p.db.Exec("insert into data (id, short, long, deleted)  values ($1, $2, $3, false)", note.ID, note.Short, note.Long)
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

func (p postgresDatabase) WriteBatch(data []model.DataBatchItem) error {
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
	rows, err := p.db.Query("select short, long from data where id=$1", id)
	if err != nil {
		log.Println(err)
		return map[string]string{}, err
	}

	result := map[string]string{}
	for rows.Next() {
		var short, long string
		err = rows.Scan(&short, &long)
		if err != nil {
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

func (p postgresDatabase) Delete(note model.Note) error {
	_, err := p.db.Exec("update data set deleted = $1 where id = $2 and short = $3", true, note.ID, note.Short)
	if err != nil {
		return err
	}

	return nil
}

func (p postgresDatabase) Start() error {
	_, err := p.db.Exec("create table data (id varchar(30), short varchar(60) UNIQUE, long text, correlation varchar(30), deleted bool)")
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

func (p postgresDatabase) Ping() error {
	return p.db.Ping()
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
