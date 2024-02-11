package handlers

import (
	"myapp/data"
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

func (h *Handlers) Form(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)

	validator := h.App.Validator(nil)

	vars.Set("validator", validator)
	vars.Set("user", data.User{})

	if err := h.render(w, r, "form", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) SubmitForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.ErrorLog.Println(err)
	}

	validator := h.App.Validator(nil)

	validator.Required(r, "first_name", "last_name", "email")
	validator.IsEmail("email", r.Form.Get("email"))
	validator.Check(len(r.Form.Get("first_name")) > 1,
		"first_name", "Must be at least 2 characters")
	validator.Check(len(r.Form.Get("last_name")) > 1,
		"last_name", "Must be at least 2 characters")

	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		var user data.User
		user.FirstName = r.Form.Get("first_name")
		user.LastName = r.Form.Get("last_name")
		user.Email = r.Form.Get("email")
		vars.Set("user", user)

		if err := h.render(w, r, "form", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
	}
}
