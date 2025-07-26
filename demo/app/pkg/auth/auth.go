package auth

import (
	"app/pkg"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	cfg *pkg.Config
}

func NewService() *Service {
	cfg := pkg.NewConfig()
	return &Service{cfg: cfg}
}

const (
	GetNotes   int64 = 0x0000000000000001
	CreateNote int64 = 0x0000000000000002
	EditNote   int64 = 0x0000000000000004
	RemoveNote int64 = 0x0000000000000008

	GetPayments   int64 = 0x0000000000000010
	CreatePayment int64 = 0x0000000000000020

	GetEmails int64 = 0x0000000000000040
	SendEmail int64 = 0x0000000000000080

	GetFiles     int64 = 0x0000000000000100
	UploadFile   int64 = 0x0000000000000200
	DownloadFile int64 = 0x0000000000000400
	RemoveFile   int64 = 0x0000000000000800

	// Admin access
	GetUsers int64 = 0x0000000000001000
	EditUser int64 = 0x0000000000002000

	// Plan access
	BasicPlan   int64 = 0x0000000000004000
	PremiumPlan int64 = 0x0000000000008000
)

const UserAccess int64 = GetNotes |
	CreateNote |
	EditNote |
	RemoveNote |
	GetPayments |
	CreatePayment |
	GetEmails |
	SendEmail |
	GetFiles |
	UploadFile |
	DownloadFile |
	RemoveFile

const AdminAccess int64 = UserAccess |
	GetUsers |
	EditUser

const NewUserAccess int64 = AdminAccess

func (s *Service) Auth(token string, access int64) (*AccessTokenClaims, error) {
	user, err := s.ValidateAccessToken(token)
	if err != nil {
		return nil, pkg.UnauthorizedError{Err: fmt.Errorf("error validating access token: %w", err)}
	}
	if !s.HasAccess(access, user.Access) {
		return nil, pkg.ForbiddenError{Err: fmt.Errorf("user does not have access to this resource")}
	}
	return user, nil
}

// Basic RBAC (Role-Based Access Control) to check if the user has the right access to the resource
func (s *Service) HasAccess(access int64, userAccess int64) bool {
	if userAccess == 0 {
		return false
	}
	return userAccess&access == access
}

func (s *Service) UpdateAccess(userAccess int64, access int64) (int64, error) {
	userAccess |= access
	return userAccess, nil
}

// Here you can implement ABAC (Attribute-Based Access Control) to check if the user has the right attributes to access the resource
type UserAttr struct {
	Department string
	Position   string
}

func (s *Service) HasAccessABAC(access int64, userAccess int64, userAttr UserAttr) bool {
	if userAccess == 0 {
		return false
	}
	if userAccess&access != access {
		return false
	}
	return CheckUserAttr(access, &userAttr)
}

func CheckUserAttr(access int64, userAttr *UserAttr) bool {
	if access == EditNote {
		return userAttr.Department == "IT" && userAttr.Position == "Developer"
	}
	if access == RemoveNote {
		return userAttr.Department == "IT" && userAttr.Position == "Admin"
	}
	return true
}

func (s *Service) GenerateTokens(
	refreshTokenID string,
	userID string,
	access int64,
	avatar string,
	email string,
	subscriptionActive bool,
) (string, string, error) {
	// Load the private key
	privateKey, err := os.ReadFile("/private.pem")
	if err != nil {
		return "", "", fmt.Errorf("error reading private key: %w", err)
	}
	if len(privateKey) == 0 {
		return "", "", fmt.Errorf("error reading private key: %w", err)
	}
	// Parse the private key
	privateKeyParsed, err := jwt.ParseEdPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("error parsing private key: %w", err)
	}
	// Create the token
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"id":                  userID,
		"access":              access,
		"avatar":              avatar,
		"email":               email,
		"subscription_active": subscriptionActive,
		"exp":                 time.Now().Add(s.cfg.AccessTokenExp).Unix(),
	})
	// Sign the token
	tokenString, err := token.SignedString(privateKeyParsed)
	if err != nil {
		return "", "", fmt.Errorf("error signing token: %w", err)
	}
	// Create the refresh token
	refreshToken := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"id":      refreshTokenID,
		"user_id": userID,
		"exp":     time.Now().Add(s.cfg.RefreshTokenExp).Unix(),
	})
	// Sign the refresh token
	refreshTokenString, err := refreshToken.SignedString(privateKeyParsed)
	if err != nil {
		return "", "", fmt.Errorf("error signing refresh token: %w", err)
	}
	return tokenString, refreshTokenString, nil
}

type AccessTokenClaims struct {
	ID                 uuid.UUID `json:"id"`
	Access             int64     `json:"access"`
	Avatar             string    `json:"avatar"`
	Email              string    `json:"email"`
	SubscriptionActive bool      `json:"subscription_active"`
}

func (s *Service) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims, err := extractTokenClaims(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	accessTokenClaims, err := extractAccessTokenClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	return accessTokenClaims, nil
}

type RefreshTokenClaims struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (s *Service) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	claims, err := extractTokenClaims(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	refreshTokenClaims, err := extractRefreshTokenClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	return refreshTokenClaims, nil
}

func extractTokenClaims(tokenString string) (jwt.MapClaims, error) {
	// Check if token starts with "Bearer "
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = tokenString[7:]
	}
	// Load the public key
	publicKey, err := os.ReadFile("/public.pem")
	if err != nil {
		return nil, fmt.Errorf("error reading public key: %w", err)
	}
	if len(publicKey) == 0 {
		return nil, fmt.Errorf("error reading public key: %w", err)
	}
	// Parse the public key
	publicKeyParsed, err := jwt.ParseEdPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}
	// Parse the token
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
		return publicKeyParsed, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	// Validate the refresh token
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}
	// Get the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("error getting claims")
	}
	return claims, nil
}

var ErrInvalidClaims = errors.New("claims are invalid")

func extractRefreshTokenClaims(claims jwt.MapClaims) (*RefreshTokenClaims, error) {
	id, ok := claims["id"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'id' field: %w", ErrInvalidClaims)
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'user_id' field: %w", ErrInvalidClaims)
	}
	UUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing user ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("error parsing refresh token ID: %w", err)
	}
	refreshTokenClaims := RefreshTokenClaims{
		ID:     UUID,
		UserID: userUUID,
	}
	return &refreshTokenClaims, nil
}

func extractAccessTokenClaims(claims jwt.MapClaims) (*AccessTokenClaims, error) {
	id, ok := claims["id"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'id' field: %w", ErrInvalidClaims)
	}
	access, ok := claims["access"].(float64)
	if !ok {
		return nil, fmt.Errorf("claims missing 'access' field: %w", ErrInvalidClaims)
	}
	avatar, ok := claims["avatar"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'avatar' field: %w", ErrInvalidClaims)
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'email' field: %w", ErrInvalidClaims)
	}
	subscriptionActive, ok := claims["subscription_active"].(bool)
	if !ok {
		return nil, fmt.Errorf("claims missing 'subscription_active' field: %w", ErrInvalidClaims)
	}
	UUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing user ID: %w", err)
	}
	accessTokenClaims := AccessTokenClaims{
		ID:                 UUID,
		Access:             int64(access),
		Avatar:             avatar,
		Email:              email,
		SubscriptionActive: subscriptionActive,
	}
	return &accessTokenClaims, nil
}

type SessionTokenClaims struct {
	ID    uuid.UUID `json:"id"`
	Phone string    `json:"phone"`
}

func (s *Service) GenerateSessionToken(
	userID string,
	phone string,
) (string, error) {
	// Load the private key
	privateKey, err := os.ReadFile("/private.pem")
	if err != nil {
		return "", fmt.Errorf("error reading private key: %w", err)
	}
	if len(privateKey) == 0 {
		return "", fmt.Errorf("error reading private key: %w", err)
	}
	// Parse the private key
	privateKeyParsed, err := jwt.ParseEdPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", fmt.Errorf("error parsing private key: %w", err)
	}
	// Create the token
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"id":    userID,
		"phone": phone,
		"exp":   time.Now().Add(s.cfg.AccessTokenExp).Unix(),
	})
	// Sign the token
	tokenString, err := token.SignedString(privateKeyParsed)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}

func (s *Service) ValidateSessionToken(tokenString string) (*SessionTokenClaims, error) {
	claims, err := extractTokenClaims(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	sessionTokenClaims, err := extractSessionTokenClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	return sessionTokenClaims, nil
}

func extractSessionTokenClaims(claims jwt.MapClaims) (*SessionTokenClaims, error) {
	id, ok := claims["id"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'id' field: %w", errors.New("invalid claims"))
	}
	UUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing user ID: %w", err)
	}
	phone, ok := claims["phone"].(string)
	if !ok {
		return nil, fmt.Errorf("claims missing 'phone' field: %w", errors.New("invalid claims"))
	}

	sessionTokenClaims := SessionTokenClaims{
		ID:    UUID,
		Phone: phone,
	}
	return &sessionTokenClaims, nil
}

