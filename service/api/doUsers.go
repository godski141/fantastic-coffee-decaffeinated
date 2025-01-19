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

    // Controllo se l'utente esiste nel database
    _, err := rt.db.GetUserByID(userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

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
        photoPath = "service/uploads/default_user_photo.jpg" // Percorso della foto predefinita
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

    lowername := strings.ToLower(req.Name)
    // Verifico che il nome non esista già
    _, err = rt.db.GetUserByName(lowername)
    if err == nil {
        http.Error(w, "Name already exists", http.StatusBadRequest)
        return
    }

    // Modifica il nome dell'utente
    err = rt.db.ModifyUserName(userID, lowername)
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

    // Rimuovi il prefisso se presente tra PNG e JPEG
    photoData := strings.TrimPrefix(req.PhotoBase64, "data:image/png;base64,")
    photoData = strings.TrimPrefix(photoData, "data:image/jpeg;base64,")

    // Decodifica l'immagine Base64
    decodedPhoto, err := base64.StdEncoding.DecodeString(photoData)
    if err != nil {
        http.Error(w, "Invalid Base64 encoding", http.StatusBadRequest)
        return
    }

    // Percorso per la directory e il file
	dirPath := "service/uploads/users/"
	filePath := fmt.Sprintf("%s%s_photo.png", dirPath, userID)

	// Assicurati che la directory esista
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Salva l'immagine nel file
	err = os.WriteFile(filePath, decodedPhoto, 0644)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

    // Aggiorna il percorso della foto nel database
    err = rt.db.UpdateUserPhoto(userID, filePath)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Risposta
    w.WriteHeader(http.StatusNoContent)
}

