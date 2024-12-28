package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
	"testgenerate_backend_subject/internal/app"
)

var (
	ErrBadRouting           = errors.New("inconsistent mapping between route and handler (programmer error)")
	ErrNotFound             = errors.New("not found")
	ErrAlreadyExists        = errors.New("this row is already exists")
	ErrInconsistentIDs      = errors.New("inconsistent IDs")
	ErrForbidden            = errors.New("role is not administrator")
	ErrPreconditionRequired = errors.New("header get authorization")
)

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, Accept, Content-Type, Content-Length, Accept-Encoding")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func MakeHTTPHandler(s Service, logger UnitLogHandler) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("OPTIONS", "GET").Path("/subjects").Handler(accessControl(httptransport.NewServer(
		e.GetSubjectsEndpoint,
		decodeGetSubjectsRequest,
		encodeResponse,
		options...,
	)))

	r.Methods("OPTIONS", "POST").Path("/subject").Handler(accessControl(httptransport.NewServer(
		e.PostSubjectEndpoint,
		decodePostSubjectRequest,
		encodeResponse,
		options...,
	)))

	r.Methods("OPTIONS", "PUT").Path("/subject").Handler(accessControl(httptransport.NewServer(
		e.PutSubjectEndpoint,
		decodePutSubjectRequest,
		encodeResponse,
		options...,
	)))

	r.Methods("OPTIONS", "DELETE").Path("/subject/{id}").Handler(accessControl(httptransport.NewServer(
		e.DeleteSubjectEndpoint,
		decodeDeleteSubjectRequest,
		encodeResponse,
		options...,
	)))

	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	return r
}

// ----------------------------------------------------------------------------------------------------------------------
func decodeGetSubjectsRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	_, role, errToken := getPermissionParams(r, err)
	if errToken != nil {
		return nil, errToken
	}
	if strings.ToLower(role) != "administrator" {
		return nil, ErrForbidden
	}
	return getSubjectsRequest{}, nil
}

func decodePostSubjectRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var addSubject app.Subject
	if e := json.NewDecoder(r.Body).Decode(&addSubject); e != nil {
		return nil, e
	}
	return postSubjectRequest{addSubject}, nil
}

func decodePutSubjectRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	_, role, errToken := getPermissionParams(r, err)
	if errToken != nil {
		return nil, errToken
	}
	if strings.ToLower(role) != "administrator" {
		return nil, ErrForbidden
	}

	var updateSubject app.Subject
	if e := json.NewDecoder(r.Body).Decode(&updateSubject); e != nil {
		return nil, e
	}

	return putSubjectRequest{updateSubject}, nil
}

func decodeDeleteSubjectRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	_, role, errToken := getPermissionParams(r, err)
	if errToken != nil {
		return nil, errToken
	}
	if strings.ToLower(role) != "administrator" {
		return nil, ErrForbidden
	}

	vars := mux.Vars(r)
	value, ok := vars["id"]
	id, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrBadRouting
	}
	return deleteSubjectRequest{id}, nil
}

// ---------------------------------------------------------------------------------------------------------------------
type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
func codeFrom(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists), errors.Is(err, ErrInconsistentIDs):
		return http.StatusBadRequest
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrPreconditionRequired):
		return http.StatusPreconditionRequired
	default:
		return http.StatusInternalServerError
	}
}

// ----------------------------------------------------------------------------------------------------------------------
func getPermissionParams(r *http.Request, err error) (string, string, error) {
	tb := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tb) != 2 {
		return "", "", ErrPreconditionRequired
	}
	user, role, err := extractTokenMetadata(tb[1])
	if err != nil {
		return "", "", ErrPreconditionRequired
	}
	return user, role, nil
}

func extractTokenMetadata(headerToken string) (string, string, error) {
	var user, role string
	token, err := jwt.Parse(headerToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Printf("extractTokenMetadata unexpected signing method: %v", token.Header["alg"])
			return nil, errors.New("unexpected signing method")
		}
		return []byte(app.GetEnv("SECRET_KEY", "secretkey")), nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user, ok = claims["username"].(string)
		if !ok {
			fmt.Printf("extractTokenMetadata not username")
			return "", "", errors.New("not username")
		}
		role, ok = claims["role"].(string)
		if !ok {
			fmt.Printf("extractTokenMetadata not role")
			return "", "", errors.New("not role")
		}
	}
	return user, role, nil
}
