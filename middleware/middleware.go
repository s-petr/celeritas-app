package middleware

import (
	"myapp/data"

	"github.com/s-petr/celeritas"
)

type Middleware struct {
	App    *celeritas.Celeritas
	Models *data.Models
}
