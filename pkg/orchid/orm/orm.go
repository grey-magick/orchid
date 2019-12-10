package orm

import (
	_ "github.com/lib/pq"
)

type ORM struct {
}

func NewORM() *ORM {
	return &ORM{}
}
