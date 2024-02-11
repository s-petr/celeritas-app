package handlers

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "cache", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	if err := h.App.ReadJSON(w, r, &userInput); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.ErrorUnauthorized401(w, r)
		return
	}

	if err := h.App.Cache.Set(userInput.Name, userInput.Value); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "saved in cache"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)

}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {
	var msg string
	var inCache = true

	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	if err := h.App.ReadJSON(w, r, &userInput); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.ErrorUnauthorized401(w, r)
		return
	}

	fromCache, err := h.App.Cache.Get(userInput.Name)
	if err != nil {
		msg = "not found in cache"
		inCache = false
	}

	var resp struct {
		Error   bool   `json:"error"`
		Value   string `json:"value"`
		Message string `json:"message"`
	}

	if inCache {
		resp.Error = false
		resp.Message = "retrieved from cache"
		resp.Value = fromCache.(string)
	} else {
		resp.Error = true
		resp.Message = msg
	}

	_ = h.App.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	if err := h.App.ReadJSON(w, r, &userInput); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.ErrorUnauthorized401(w, r)
		return
	}

	if err := h.App.Cache.Forget(userInput.Name); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "deleted from cache (if existed)"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		CSRF string `json:"csrf_token"`
	}

	if err := h.App.ReadJSON(w, r, &userInput); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.ErrorUnauthorized401(w, r)
		return
	}

	if err := h.App.Cache.Empty(); err != nil {
		h.App.ErrorIntServErr500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "emptied all data from cache"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)
}
