package handler

import (
	"00-go-base-tpl-sv/internal/player"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/mymmrac/telego"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"os"
)

type playerRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type errorResponse struct {
	Error error `json:"error"`
}

type playerResponse struct {
	Player player.Player `json:"player"`
}

type playersResponse struct {
	Players []player.Player `json:"players"`
}

type Players struct {
	service player.Service
	logger  *zap.Logger
}

type Question struct {
	Text               string `json:"text" validate:"required"`
	Option1            string `json:"option_1" validate:"required"`
	Option2            string `json:"option_2" validate:"required"`
	Option3            string `json:"option_3" validate:"required"`
	Option4            string `json:"option_4" validate:"required"`
	CorrectOptionValue string `json:"correct_option" validate:"required"`
}

func NewPlayers(service player.Service, logger *zap.Logger) *Players {
	return &Players{service: service, logger: logger}
}

func (h *Players) Register(r *mux.Router) {
	r.HandleFunc("/", h.app).Name("home").Methods("GET")
	r.HandleFunc("/questions", h.getQuestions).Name("questions").Methods("GET")

	r.HandleFunc("/me", h.me).Name("me").Methods("GET")
	r.HandleFunc("/callback", h.handleCallback).Name("callback").Methods("GET")

	r.HandleFunc("/players", h.list).Name("list_players").Methods("GET")
	r.HandleFunc("/players", h.create).Name("create_player").Methods("POST")
	r.HandleFunc("/players/filter", h.filter).Name("filter_players").Methods("GET")
	r.HandleFunc("/players/{id}", h.read).Name("read_player").Methods("GET")
	r.HandleFunc("/players/{id}", h.update).Name("update_player").Methods("PATCH", "PUT")
	r.HandleFunc("/players/{id}", h.delete).Name("delete_player").Methods("DELETE")
}

func (h *Players) writeServiceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, player.ErrNotFound):
		h.writeErr(
			w,
			err,
			http.StatusNotFound,
		)
	case errors.Is(err, player.ErrConflict):
		h.writeErr(
			w,
			err,
			http.StatusConflict,
		)
	case errors.Is(err, player.ErrVersionMismatch):
		h.writeErr(
			w,
			err,
			http.StatusConflict,
		)
	default:
		h.writeErr(
			w,
			err,
			http.StatusInternalServerError,
		)
	}
}

func (h *Players) writeErr(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	h.writeResponse(w, errorResponse{Error: err})
}

func (h *Players) writeResponse(w http.ResponseWriter, v interface{}) {
	data, err := sonic.ConfigFastest.Marshal(v)
	if err != nil {
		h.logger.Error("json marshal error", zap.Error(err))
		return
	}

	_, err = w.Write(data)
	if err != nil {
		h.logger.Error("write error", zap.Error(err))
	}
}

func (h *Players) create(w http.ResponseWriter, r *http.Request) {
	var req playerRequest

	dec := sonic.ConfigFastest.NewDecoder(r.Body)
	err := dec.Decode(&req)

	if err != nil {
		h.writeErr(w, fmt.Errorf("unmarshal request body: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	p, err := h.service.Create(ctx, req.Email, req.Name)
	if err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, playerResponse{Player: p})
}

func (h *Players) read(w http.ResponseWriter, r *http.Request) {
	id, err := xid.FromString(mux.Vars(r)["id"])
	if err != nil {
		h.writeErr(w, fmt.Errorf("parse id: %w", err), http.StatusBadRequest)
		return
	}

	p, err := h.service.Read(r.Context(), id)
	if err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, playerResponse{Player: p})
}

func (h *Players) update(w http.ResponseWriter, r *http.Request) {
	dec := sonic.ConfigFastest.NewDecoder(r.Body)
	var req playerRequest

	err := dec.Decode(&req)
	if err != nil {
		h.writeErr(w, fmt.Errorf("unmarshal request body: %w", err), http.StatusBadRequest)
		return
	}

	id, err := xid.FromString(mux.Vars(r)["id"])
	if err != nil {
		h.writeErr(w, fmt.Errorf("parse id: %w", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	p, err := h.service.Update(ctx, id, req.Email, req.Name)
	if err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, playerResponse{Player: p})
}

func (h *Players) delete(w http.ResponseWriter, r *http.Request) {
	id, err := xid.FromString(mux.Vars(r)["id"])
	if err != nil {
		h.writeErr(w, err, http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, struct{}{})
}

func (h *Players) home(w http.ResponseWriter, r *http.Request) {
	h.writeResponse(w, "ok")
}

func (h *Players) getQuestions(w http.ResponseWriter, r *http.Request) {
	q1 := Question{"What hero has finger?", "Lina", "Lion", "Templar Assasin", "Spirit braker", "Lion"}
	q2 := Question{"What can dive?", "Pudge", "Io", "Ember spirit", "Phoenix", "Phoenix"}

	questions := []Question{q1, q2}

	h.writeResponse(w, questions)
}

func (h *Players) app(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML file
	tmpl, err := template.ParseFiles("../../internal/frontend/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Data to pass into the template
	data := struct {
		Name string
	}{
		Name: "Variable!@#123123",
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Players) handleCallback(w http.ResponseWriter, r *http.Request) {
	botToken := "6484612269:AAGMCqUOTmfDI4KDPfumW3hZO3s0LmcoozQ" // todo

	// Note: Please keep in mind that default logger may expose sensitive information,
	// use in development only
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get updates channel
	// (more on configuration in examples/updates_long_polling/main.go)
	updates, _ := bot.UpdatesViaLongPolling(nil)

	// Stop reviving updates from update channel
	defer bot.StopLongPolling()

	// Loop through all updates when they came
	for update := range updates {
		fmt.Printf("Update: %+v\n", update)
	}

	h.writeResponse(w, updates)
}
func (h *Players) me(w http.ResponseWriter, r *http.Request) {
	// Get Bot token from environment variables
	botToken := "6484612269:AAGMCqUOTmfDI4KDPfumW3hZO3s0LmcoozQ" // todo

	// Create bot and enable debugging info
	// Note: Please keep in mind that default logger may expose sensitive information,
	// use in development only
	// (more on configuration in examples/configuration/main.go)
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Call method getMe (https://core.telegram.org/bots/api#getme)
	botUser, err := bot.GetMe()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Print Bot information
	fmt.Printf("Bot user: %+v\n", botUser)

	h.writeResponse(w, botUser)
}

func (h *Players) list(w http.ResponseWriter, r *http.Request) {
	pp, err := h.service.List(r.Context())
	if err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, playersResponse{Players: pp})
}

type filterReq struct {
	Name   string `schema:"name"`
	Email  string `schema:"email"`
	Offset uint   `schema:"offset"`
	Limit  uint   `schema:"limit"`
}

type filterResponse struct {
	Total   uint            `json:"total"`
	Players []player.Player `json:"players"`
}

func (h *Players) filter(w http.ResponseWriter, r *http.Request) {
	var req filterReq

	err := schema.NewDecoder().Decode(&req, r.URL.Query())
	if err != nil {
		h.writeErr(w, fmt.Errorf("decode query: %w", err), http.StatusBadRequest)
		return
	}

	total, pp, err := h.service.Filter(
		r.Context(),
		player.FilterRequest{
			Name:  req.Name,
			Email: req.Email,
		},
		req.Offset,
		req.Limit,
	)
	if err != nil {
		h.writeServiceErr(w, err)
		return
	}

	h.writeResponse(w, filterResponse{Total: total, Players: pp})
}
