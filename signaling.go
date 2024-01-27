package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pion/webrtc/v4"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

func jsonError(desc string, details ...string) []byte {
	if len(details) > 0 {
		return []byte(fmt.Sprintf("{\"error\":\"%s\",\"details\":\"%s\"}", desc, strings.Join(details, "\n")))
	} else {
		return []byte(fmt.Sprintf("{\"error\":\"%s\"}", desc))
	}
}

func loadJwtPublicKey(config *JwtRestreamConfig) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(config.Public))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := publicKey.(type) {
	case *ecdsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("unsupported key type, only ECDSA is supported")
	}
}

type handler struct {
	config *RestreamConfig
	tracks *TrackSet
	jwtKey func(token *jwt.Token) (interface{}, error)
}

func (h *handler) verifyToken(token string) error {
	_, err := jwt.Parse(token, h.jwtKey)

	if err != nil {
		return err
	}

	return nil
}

type ExchangeRequest struct {
	Sid         string                    `json:"sid"`
	Description webrtc.SessionDescription `json:"description"`
	Candidates  []webrtc.ICECandidateInit `json:"candidates"`
}

type ExchangeResponse struct {
	Description *webrtc.SessionDescription `json:"description"`
	Candidates  []*webrtc.ICECandidateInit `json:"candidates"`
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == "POST" {
		token := r.Header.Get("Authorization")

		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonError("auth_required"))
			return
		}

		if !strings.HasPrefix(token, bearerPrefix) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonError("invalid_token"))
			return
		}

		token = token[len(bearerPrefix):]

		validOrErr := h.verifyToken(token)

		if validOrErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(jsonError("unauthorized", validOrErr.Error()))
			return
		}

		var body ExchangeRequest

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonError("invalid_body", err.Error()))
			return
		}

		track := h.tracks.Get(body.Sid)
		if track == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(jsonError("track_not_found"))
			return
		}

		answer, candidates, err := HandlePeer(body.Description, &h.config.Ice, track, body.Candidates)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(jsonError("internal_error", err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ExchangeResponse{Description: answer, Candidates: candidates})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(jsonError("unhandled_method"))
	return
}

type SignalingState struct {
	Config *RestreamConfig
	Tracks *TrackSet
}

func Serve(state SignalingState) {
	pubkey, jwtErr := loadJwtPublicKey(&state.Config.Jwt)

	if jwtErr != nil {
		panic(jwtErr)
	}

	app := http.NewServeMux()

	app.Handle("/", &handler{
		config: state.Config,
		tracks: state.Tracks,
		jwtKey: func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return pubkey, nil
		},
	})

	sigconf := state.Config.Signaling
	addr := fmt.Sprintf("%s:%d", sigconf.Host, sigconf.Port)
	err := http.ListenAndServe(addr, app)

	if err != nil {
		panic(err)
	}
}
