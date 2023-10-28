package example

import (
	"github.com/cohix/libsdk/pkg/store"
	"github.com/pkg/errors"
)

type Person struct {
	ID        int64  `db:"person_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}

var Migrations = []string{m1}

const m1 = `
CREATE TABLE people (
	person_id INTEGER PRIMARY KEY,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE
);
`

func InsertPersonHandler(tx store.Tx, args ...any) (any, error) {
	q := `
	INSERT INTO people (first_name, last_name, email)
	VALUES($1, $2, $3);
	`

	id, err := tx.ReadWrite().Exec(q, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Exec")
	}

	return id, nil
}

func GetPersonHandler(tx store.Tx, args ...any) (any, error) {
	q := `
	SELECT
		person_id,
		first_name,
		last_name,
		email
	FROM
		people
	WHERE
		person_id=$1;
	`

	person := &Person{}

	if err := tx.Read().Get(person, q, args...); err != nil {
		return nil, errors.Wrap(err, "failed to Get")
	}

	return person, nil
}
