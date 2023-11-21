package main

import (
	"github.com/cohix/libsdk/pkg/store"
	"github.com/pkg/errors"
)

type Person struct {
	PersonID  int64  `db:"person_id" json:"person_id"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName  string `db:"last_name" json:"last_name"`
	Email     string `db:"email" json:"email"`
}

func personSvcMigrations() []string {
	return []string{m1}
}

const m1 = `
CREATE TABLE people (
	person_id INTEGER PRIMARY KEY,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE
);
`

// InsertPerson and InsertPersonTx are an example of the best practice for defining a transaction handler.
// By defining a TxName variable and TxHandler func on the same line, it makes Jump-To-Definition more useful.
// and the codebase easier to reason about. You can also define them on seperate lines if you prefer (see below).
var InsertPerson, InsertPersonTx = store.TxName("InsertPerson"), func(tx store.Tx, args ...any) (any, error) {
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

// SelectPeople and SelectPeopleTx below is the more verbose two-line way to define transactions.
// The above one-line version is preferable, but this is also acceptable!
var SelectPeople = store.TxName("SelectPeople")

func SelectPeopleTx(tx store.Tx, args ...any) (any, error) {
	q := `
	SELECT
		person_id,
		first_name,
		last_name,
		email
	FROM
		people
	LIMIT 10;
	`

	ppl := []Person{}

	if err := tx.Read().Select(&ppl, q, args...); err != nil {
		return nil, errors.Wrap(err, "failed to Get")
	}

	return ppl, nil
}

var GetPerson, GetPersonTx = store.TxName("GetPerson"), func(tx store.Tx, args ...any) (any, error) {
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
