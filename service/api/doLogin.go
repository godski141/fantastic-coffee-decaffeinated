package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

type LoginRequest struct {
    Name string `json:"username"`
}

type LoginResponse struct {
    Identifier string `json:"userID"`
}


// doLogin handles POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

    // Decodifica il corpo della richiesta
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Verifica che il campo Name non sia vuoto
    if req.Name == "" {
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }

    // Controlla se l'utente esiste nel database
    id, err := rt.db.GetUserByName(req.Name)
    if err != nil {

        // Se l'utente non esiste, crea un nuovo utente
        id, err = rt.db.CreateUser(req.Name)
        if err != nil {

            // Se c'è un errore nella creazione dell'utente, ritorna errore
            http.Error(w, "Error creating user", http.StatusInternalServerError)
            return
        }
    }
    
    // Invia l'ID dell'utente come risposta
    res := LoginResponse{Identifier: id}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}