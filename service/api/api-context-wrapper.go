package api

import (
	"WasaTEXT/service/api/reqcontext"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// ExtractIdFromBearer estrae l'ID utente dal token Bearer
func ExtractIdFromBearer(token string) (string, error) {
	if len(token) < len("Bearer ") || token[:len("Bearer ")] != "Bearer " {
		return "", errors.New("invalid Bearer token format")
	}
	return token[len("Bearer "):], nil
}

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
func (rt *_router) wrap(fn httpRouterHandler, authRequired bool) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID,
		}

		// Create a request-specific logger
		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
		})

		// Autenticazione
		if authRequired {
			ctx.UserId, err = ExtractIdFromBearer(r.Header.Get("Authorization"))
			if err != nil {
				ctx.Logger.WithError(err).Error("invalid Bearer token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Check if the user exists
			_, err = rt.db.GetUserByID(ctx.UserId)
			if err != nil {
				ctx.Logger.WithError(err).Error("User not found")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		// Call the next handler in chain (usually, the handler function for the path)
		fn(w, r, ps, ctx)
	}
}
