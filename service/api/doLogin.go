package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type LoginRequest struct {
    Name string `json:"name"`
}

type LoginResponse struct {
    Identifier string `json:"identifier"`
}


// doLogin handles POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    log.Println("DEBUG: Received login request")

    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Println("ERROR: Failed to decode request body:", err)
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    log.Println("DEBUG: Request body decoded successfully. Name:", req.Name)

    // Check if name is empty
    if req.Name == "" {
        log.Println("ERROR: Name cannot be empty")
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }

    log.Println("DEBUG: Checking if user exists in database")

    // Check if user exists
    id, err := rt.db.GetUserByName(req.Name)
    if err != nil {
        log.Println("DEBUG: User not found. Attempting to create new user...")

        // If user does not exist, create a new user
        id, err = rt.db.CreateUser(req.Name)
        if err != nil {
            log.Println("ERROR: Failed to create user in database:", err)
            http.Error(w, "Error creating user", http.StatusInternalServerError)
            return
        }

        log.Println("DEBUG: New user created successfully. ID:", id)
    } else {
        log.Println("DEBUG: User already exists. ID:", id)
    }
    
    // Respond with the user ID
    res := LoginResponse{Identifier: id}
    w.Header().Set("Content-Type", "application/json")

    log.Println("DEBUG: Sending response with user ID:", id)
    json.NewEncoder(w).Encode(res)
}