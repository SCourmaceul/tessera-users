// Package http contient les handlers lambdahttp du service : requête de la gateway → cas
// d'usage → réponse JSON. Routes déclarées dans terraform/main.tf.
package http

import (
	"context"
	"time"

	"github.com/SCourmaceul/tetra-kit/adapter/lambdahttp"

	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// meResponse est le profil de l'utilisateur connecté (GET/PATCH /api/v1/users/me).
type meResponse struct {
	ID        string               `json:"id"`
	Pseudo    string               `json:"pseudo"`
	AvatarURL string               `json:"avatar_url"`
	CreatedAt time.Time            `json:"created_at"`
	Spaces    []membershipResponse `json:"spaces"`
}

type membershipResponse struct {
	Space      string    `json:"space"`
	Role       string    `json:"role"`
	Reputation int       `json:"reputation"`
	JoinedAt   time.Time `json:"joined_at"`
}

// userResponse est un utilisateur vu depuis un espace.
type userResponse struct {
	ID         string    `json:"id"`
	Pseudo     string    `json:"pseudo"`
	AvatarURL  string    `json:"avatar_url"`
	Space      string    `json:"space"`
	Role       string    `json:"role"`
	Reputation int       `json:"reputation"`
	JoinedAt   time.Time `json:"joined_at"`
}

type listResponse struct {
	Users []userResponse `json:"users"`
}

type updateRequest struct {
	Pseudo    *string `json:"pseudo"`
	AvatarURL *string `json:"avatar_url"`
}

func toMeResponse(me app.Me) meResponse {
	resp := meResponse{
		ID:        me.Profile.ID,
		Pseudo:    me.Profile.Pseudo,
		AvatarURL: me.Profile.AvatarURL,
		CreatedAt: me.Profile.CreatedAt,
		Spaces:    make([]membershipResponse, len(me.Members)),
	}
	for i, m := range me.Members {
		resp.Spaces[i] = membershipResponse{Space: m.Space, Role: m.Role, Reputation: m.Reputation, JoinedAt: m.JoinedAt}
	}
	return resp
}

func toUserResponse(m domain.Member) userResponse {
	return userResponse{
		ID:         m.UserID,
		Pseudo:     m.Pseudo,
		AvatarURL:  m.AvatarURL,
		Space:      m.Space,
		Role:       m.Role,
		Reputation: m.Reputation,
		JoinedAt:   m.JoinedAt,
	}
}

// GetMe : GET /api/v1/users/me
func GetMe(uc app.GetMe) lambdahttp.HandlerFunc {
	return func(ctx context.Context, req *lambdahttp.Request) (lambdahttp.Response, error) {
		userID, err := req.RequireUser()
		if err != nil {
			return lambdahttp.Response{}, err
		}
		me, err := uc.Execute(ctx, userID)
		if err != nil {
			return lambdahttp.Response{}, err
		}
		return lambdahttp.OK(toMeResponse(me))
	}
}

// UpdateMe : PATCH /api/v1/users/me, corps {"pseudo"?, "avatar_url"?}
func UpdateMe(uc app.UpdateMe) lambdahttp.HandlerFunc {
	return func(ctx context.Context, req *lambdahttp.Request) (lambdahttp.Response, error) {
		userID, err := req.RequireUser()
		if err != nil {
			return lambdahttp.Response{}, err
		}
		var body updateRequest
		if err := req.DecodeBody(&body); err != nil {
			return lambdahttp.Response{}, err
		}
		me, err := uc.Execute(ctx, userID, app.UpdateInput{Pseudo: body.Pseudo, AvatarURL: body.AvatarURL})
		if err != nil {
			return lambdahttp.Response{}, err
		}
		return lambdahttp.OK(toMeResponse(me))
	}
}

// JoinSpace : POST /api/v1/users/me/spaces, rejoint l'espace de la requête
func JoinSpace(uc app.JoinSpace) lambdahttp.HandlerFunc {
	return func(ctx context.Context, req *lambdahttp.Request) (lambdahttp.Response, error) {
		userID, err := req.RequireUser()
		if err != nil {
			return lambdahttp.Response{}, err
		}
		m, err := uc.Execute(ctx, req.Space(), userID)
		if err != nil {
			return lambdahttp.Response{}, err
		}
		return lambdahttp.Created(toUserResponse(m))
	}
}

// GetUser : GET /api/v1/users/:id
func GetUser(uc app.GetUser) lambdahttp.HandlerFunc {
	return func(ctx context.Context, req *lambdahttp.Request) (lambdahttp.Response, error) {
		m, err := uc.Execute(ctx, req.Space(), req.PathParam("id"))
		if err != nil {
			return lambdahttp.Response{}, err
		}
		return lambdahttp.OK(toUserResponse(m))
	}
}

// ListUsers : GET /api/v1/users
func ListUsers(uc app.ListUsers) lambdahttp.HandlerFunc {
	return func(ctx context.Context, req *lambdahttp.Request) (lambdahttp.Response, error) {
		members, err := uc.Execute(ctx, req.Space())
		if err != nil {
			return lambdahttp.Response{}, err
		}
		resp := listResponse{Users: make([]userResponse, len(members))}
		for i, m := range members {
			resp.Users[i] = toUserResponse(m)
		}
		return lambdahttp.OK(resp)
	}
}
