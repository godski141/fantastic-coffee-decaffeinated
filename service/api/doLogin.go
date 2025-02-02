package api

import (
	"WasaTEXT/service/api/reqcontext"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/julienschmidt/httprouter"
)

type LoginRequest struct {
    Name string `json:"username"`
}

type LoginResponse struct {
    Identifier string `json:"user_id"`
}


// doLogin handles POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {

    // Decodifica il corpo della richiesta
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        ctx.Logger.WithError(err).Error("Error decoding request")
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Verifica che il campo Name non sia vuoto
    if req.Name == "" {
        ctx.Logger.Error("Name cannot be empty")
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }

    // Verifica che il nome sia tra 3 e 50 caratteri
    if len(req.Name) < 3 || len(req.Name) > 50 {
        ctx.Logger.Error("Name must be between 3 and 50 characters")
        http.Error(w, "Name must be between 3 and 50 characters", http.StatusBadRequest)
        return
    }

    lowername := strings.ToLower(req.Name)
    // Controlla se l'utente esiste nel database
    id, err := rt.db.GetUserByName(lowername)
    if err != nil {

        // Se l'utente non esiste, crea un nuovo utente
        id, err = rt.db.CreateUser(lowername)
        if err != nil {
            ctx.Logger.WithError(err).Error("Error creating user")
            // Se c'è un errore nella creazione dell'utente, ritorna errore
            http.Error(w, "Error creating user", http.StatusInternalServerError)
            return
        }
    }
    
    ctx.Logger.WithField("user_id", id).Info("User logged in")
    // Invia l'ID dell'utente come risposta
    res := LoginResponse{Identifier: id}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}
