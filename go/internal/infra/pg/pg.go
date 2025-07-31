package pg

import (
	"github.com/doug-martin/goqu/v9"
)

var Dialect = goqu.Dialect("postgres")
