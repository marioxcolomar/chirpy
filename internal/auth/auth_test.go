package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	pass1 := "pa$$word"
	pass2 := "$ecure"
	hash1, _ := HashPassword(pass1)
	hash2, _ := HashPassword(pass2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "correct password",
			password:      pass1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "incorrect password",
			password:      "fake-password",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "password does not match",
			password:      pass1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "wrong password format",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "invalid hash",
			password:      pass1,
			hash:          "wrong-hash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("check password hash error: %v, want: %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("check password expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	t.Run("create valid token", func(t *testing.T) {
		userID := uuid.New()
		token, _ := MakeJWT(userID, "secret", time.Hour)

		tests := []struct {
			name        string
			tokenString string
			tokenSecret string
			wantUserID  uuid.UUID
			wantErr     bool
		}{
			{
				name:        "valid token",
				tokenString: token,
				tokenSecret: "secret",
				wantUserID:  userID,
				wantErr:     false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
				if (err != nil) != tt.wantErr {
					t.Errorf("validate jwt error: %v, wantErr: %v", err, tt.wantErr)
					return
				}
				if gotUserID != tt.wantUserID {
					t.Errorf("validate jwt error: %v, want: %v", gotUserID, tt.wantUserID)
				}
			})
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "correct token",
			headers: http.Header{
				"Authorization": []string{"Bearer this-is-me-token"},
			},
			wantToken: "this-is-me-token",
			wantErr:   false,
		},
		{
			name: "malformed token",
			headers: http.Header{
				"Authorization": []string{"NotBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "missing authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("get bearer token error: %v, wantErr: %v", err, tt.wantErr)
				return
			}
			if got != tt.wantToken {
				t.Errorf("get bearer token does not match got: %v want: %v", got, tt.wantToken)
				return
			}
		})
	}
}
