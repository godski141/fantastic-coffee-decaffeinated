package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type NewName struct {
    Name string `json:"new_name"`
} 

// getUserPhoto handles GET /users/get-photo/:user_id
func (rt *_router) getUserPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // Recupera l'userId dai parametri
    userID := ps.ByName("user_id")

    // Recupera il percorso della foto dal database
    photoPath, err := rt.db.GetUserPhotoByID(userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            http.Error(w, "User not found or photo not set", http.StatusNotFound)
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Se la foto non è impostata, restituisci una foto predefinita
    if photoPath == "" {
        photoPath = "uploads/default_user_photo.png" // Percorso della foto predefinita
    }

    // Serve il file immagine
    http.ServeFile(w, r, photoPath)
}

// modifyUserName handles PATCH /users/modify-name
func (rt *_router) modifyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // Recupera l'userId dal Authorization Header
	userID := r.Header.Get("Authorization")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

    // Controllo se l'utente esiste nel database
	_, err := rt.db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

    // Decodifica il corpo della richiesta
    var req NewName
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Verifica che il campo Name non sia vuoto
    if req.Name == "" {
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }

    // Verifica che il nome sia tra 3 e 50 caratteri
    if len(req.Name) < 3 || len(req.Name) > 50 {
        http.Error(w, "Name must be between 3 and 50 characters", http.StatusBadRequest)
        return
    }

    // Verifico che il nome non esista già
    _, err = rt.db.GetUserByName(req.Name)
    if err == nil {
        http.Error(w, "Name already exists", http.StatusBadRequest)
        return
    }

    // Modifica il nome dell'utente
    err = rt.db.ModifyUserName(userID, req.Name)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Username updated successfully"})
}

// updateUserPhoto handles PATCH /users/update-photo
func (rt *_router) updateUserPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // Recupero l'userId dal Authorization Header
    userID := r.Header.Get("Authorization")
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Controllo se l'utente esiste nel database
    _, err := rt.db.GetUserByID(userID)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Decodifica il corpo della richiesta
    var req struct {
        PhotoBase64 string `json:"photo"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verifica che il campo photo non sia vuoto
    if req.PhotoBase64 == "" {
        http.Error(w, "Photo cannot be empty", http.StatusBadRequest)
        return
    }

    // Rimuovi il prefisso se presente
    photoData := strings.TrimPrefix(req.PhotoBase64, "data:image/png;base64,")

    // Decodifica l'immagine Base64
    decodedPhoto, err := base64.StdEncoding.DecodeString(photoData)
    if err != nil {
        http.Error(w, "Invalid Base64 encoding", http.StatusBadRequest)
        return
    }

    photoPath := fmt.Sprintf("WasaTEXT/service/uploads/users/%s_photo.png", userID)
    err = os.WriteFile(photoPath, decodedPhoto, 0644)
    if err != nil {
        http.Error(w, "Failed to save image", http.StatusInternalServerError)
        return
    }

    // Aggiorna il percorso della foto nel database
    err = rt.db.UpdateUserPhoto(userID, photoPath)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Risposta
    w.WriteHeader(http.StatusNoContent)
}

