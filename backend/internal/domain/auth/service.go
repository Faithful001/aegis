package auth

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/domain/otp"
	"github.com/Faithful001/aegis/internal/domain/project"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/email"
	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo      user.IUserRepository
	otpRepo       otp.IOTPRepository
	otpService    otp.IOTPService
	emailService  email.IEmailService
	tokenService  IJWTService
	hasher        IBcryptPasswordHasher
	blacklistRepo ITokenBlacklistRepository
	eventProducer kafka.EventProducer
	sessionRepo   ISessionRepository
}

func NewAuthService(
	userRepo 		user.IUserRepository,
	otpRepo       	otp.IOTPRepository,
	otpService    	otp.IOTPService,
	emailService  	email.IEmailService,
	tokenService  	IJWTService,
	hasher        	IBcryptPasswordHasher,
	blacklistRepo 	ITokenBlacklistRepository,
	eventProducer ...kafka.EventProducer,
) *AuthService {
	var ep kafka.EventProducer
	if len(eventProducer) > 0 {
		ep = eventProducer[0]
	}
	return &AuthService{
		userRepo:      userRepo,
		otpRepo:       otpRepo,
		otpService:    otpService,
		emailService:  emailService,
		tokenService:  tokenService,
		hasher:        hasher,
		blacklistRepo: blacklistRepo,
		eventProducer: ep,
	}
}

func (s *AuthService) SetEventProducer(ep kafka.EventProducer) {
	s.eventProducer = ep
}

func (s *AuthService) SetSessionRepo(repo ISessionRepository) {
	s.sessionRepo = repo
}

// func (s *AuthService) SetEmailService(emailService *email.EmailService) {
// 	s.emailService = emailService
// }

// func (s *AuthService) SetOTP(otpRepo otp.OTPRepository, otpService *otp.OTPService) {
// 	s.otpRepo = otpRepo
// 	s.otpService = otpService
// }

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (bool, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return false, user.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser, err := user.NewUser(req.FirstName, req.LastName, email, hashedPassword)
	if err != nil {
		return false, err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return false, fmt.Errorf("failed to persist user: %w", err)
	}

	s.SendVerificationEmail(ctx, newUser.Email)

	return true, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	userEntity, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !userEntity.IsActive() {
		return nil, user.ErrUserInactive
	}

	if err := s.hasher.Compare(userEntity.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenPair, accessClaims, refreshClaims, err := s.tokenService.GenerateTokenPair(userEntity.ID, userEntity.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	if s.sessionRepo != nil && accessClaims != nil {
		sess := &UserSession{
			UserID:       userEntity.ID,
			AccessJTI:    accessClaims.JTI,
			RefreshJTI:   refreshClaims.JTI,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			DeviceName:   ParseDeviceName(userAgent),
			LastActiveAt: time.Now().UTC(),
			ExpiresAt:    refreshClaims.ExpiresAt,
			IsRevoked:    false,
		}
		_ = s.sessionRepo.Create(ctx, sess)
	}

	return &dto.AuthResponse{
		Token: dto.TokenResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			TokenType:    tokenPair.TokenType,
			ExpiresIn:    tokenPair.ExpiresIn,
		},
		User: userEntity,
	}, nil
}


// SendVerificationEmail sends an OTP to the authenticated user's email for verification.
func (s *AuthService) SendVerificationEmail(ctx context.Context, userEmail string) error {
	if s.emailService == nil && s.eventProducer == nil {
		return fmt.Errorf("email service is not configured")
	}
	u, err := s.userRepo.GetByEmail(ctx, userEmail)
	if err != nil {
		return user.ErrUserNotFound
	}

	if u.IsVerified {
		return ErrEmailAlreadyVerified
	}

	// Invalidate old OTPs for this user+purpose
	_ = s.otpRepo.DeleteExpiredForUser(ctx, u.ID, otp.OTPPurposeEmailVerification)

	code, err := s.otpService.GenerateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	record := &otp.OTPRecord{
		UserID:    u.ID,
		Email:     u.Email,
		Code:      code,
		Purpose:   otp.OTPPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(otp.OTPExpiry),
	}
	if err := s.otpRepo.Create(ctx, record); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	name := strings.TrimSpace(u.FirstName + " " + u.LastName)

	if s.eventProducer != nil {
		evt := kafka.NewEmailEvent("", u.Email, name, code, string(otp.OTPPurposeEmailVerification))
		if err := s.eventProducer.Publish(ctx, kafka.TopicEmailEvents, u.Email, evt); err != nil {
			log.Printf("Failed to publish verification email event to Kafka: %v", err)
			if s.emailService != nil {
				return s.emailService.SendOTPEmail(ctx, u.Email, name, code, otp.OTPPurposeEmailVerification)
			}
			return fmt.Errorf("failed to publish verification email event: %w", err)
		}
		return nil
	}

	return s.emailService.SendOTPEmail(ctx, u.Email, name, code, otp.OTPPurposeEmailVerification)
}


// VerifyEmail validates the OTP, marks the user as verified, and returns auth credentials.
func (s *AuthService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (*dto.AuthResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	if u.IsVerified {
		return nil, ErrEmailAlreadyVerified
	}

	record, err := s.otpRepo.FindValid(ctx, u.Email, req.Code, otp.OTPPurposeEmailVerification)
	if err != nil {
		return nil, ErrInvalidOTP
	}

	if err := s.otpRepo.MarkUsed(ctx, record.ID); err != nil {
		return nil, fmt.Errorf("failed to consume OTP: %w", err)
	}

	if err := s.userRepo.MarkVerified(ctx, u.ID); err != nil {
		return nil, fmt.Errorf("failed to mark user verified: %w", err)
	}

	u.IsVerified = true

	tokenPair, _, _, err := s.tokenService.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &dto.AuthResponse{
		Token: dto.TokenResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			TokenType:    tokenPair.TokenType,
			ExpiresIn:    tokenPair.ExpiresIn,
		},
		User: u,
	}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	if s.emailService == nil && s.eventProducer == nil {
		return fmt.Errorf("email service is not configured")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// silently succeed. don't reveal whether the email exists
		return nil
	}

	_ = s.otpRepo.DeleteExpiredForUser(ctx, u.ID, otp.OTPPurposePasswordReset)

	code, err := s.otpService.GenerateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	record := &otp.OTPRecord{
		UserID:    u.ID,
		Email:     u.Email,
		Code:      code,
		Purpose:   otp.OTPPurposePasswordReset,
		ExpiresAt: time.Now().UTC().Add(otp.OTPExpiry),
	}
	if err := s.otpRepo.Create(ctx, record); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	name := strings.TrimSpace(u.FirstName + " " + u.LastName)

	if s.eventProducer != nil {
		evt := kafka.NewEmailEvent("", u.Email, name, code, string(otp.OTPPurposePasswordReset))
		if err := s.eventProducer.Publish(ctx, kafka.TopicEmailEvents, u.Email, evt); err != nil {
			log.Printf("Failed to publish password reset email event to Kafka: %v", err)
			if s.emailService != nil {
				return s.emailService.SendOTPEmail(ctx, u.Email, name, code, otp.OTPPurposePasswordReset)
			}
			return fmt.Errorf("failed to publish password reset email event: %w", err)
		}
		return nil
	}

	return s.emailService.SendOTPEmail(ctx, u.Email, name, code, otp.OTPPurposePasswordReset)
}

func (s *AuthService) VerifyResetOTP(ctx context.Context, req dto.VerifyResetOTPRequest) (*dto.VerifyResetOTPResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	record, err := s.otpRepo.FindValid(ctx, email, req.Code, otp.OTPPurposePasswordReset)
	if err != nil {
		return nil, ErrInvalidOTP
	}

	if err := s.otpRepo.MarkUsed(ctx, record.ID); err != nil {
		return nil, fmt.Errorf("failed to consume OTP: %w", err)
	}

	if s.tokenService == nil {
		return nil, fmt.Errorf("token service is not configured")
	}

	resetToken, err := s.tokenService.GenerateResetPasswordToken(record.UserID, record.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	return &dto.VerifyResetOTPResponse{
		Success:    true,
		ResetToken: resetToken,
		TokenType:  string(TokenTypeResetPassword),
		Message:    "OTP verified successfully",
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token string, req dto.ResetPasswordRequest) error {
	tokenStr := strings.TrimSpace(token)
	if tokenStr == "" {
		tokenStr = strings.TrimSpace(req.Token)
	}

	var targetUserID uuid.UUID

	if tokenStr != "" {
		if s.tokenService == nil {
			return fmt.Errorf("token service is not configured")
		}
		claims, err := s.tokenService.ValidateToken(tokenStr, TokenTypeResetPassword)
		if err != nil {
			return err
		}
		targetUserID = claims.UserID
	} else if req.Email != "" && req.Code != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		record, err := s.otpRepo.FindValid(ctx, email, req.Code, otp.OTPPurposePasswordReset)
		if err != nil {
			return ErrInvalidOTP
		}
		targetUserID = record.UserID
		_ = s.otpRepo.MarkUsed(ctx, record.ID)
	} else {
		return fmt.Errorf("reset token is required")
	}

	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, targetUserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return user.ErrUserNotFound
	}

	if err := s.hasher.Compare(u.Password, req.CurrentPassword); err != nil {
		return ErrCurrentPasswordWrong
	}

	hashedPassword, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}

func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	claims, err := s.tokenService.ValidateToken(req.RefreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, ErrTokenRevoked
	}

	// Rotate refresh token
	_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeRefresh, claims.ExpiresAt, "refreshed")

	userEntity, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	if !userEntity.IsActive() {
		return nil, user.ErrUserInactive
	}

	tokenPair, _, _, err := s.tokenService.GenerateTokenPair(userEntity.ID, userEntity.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	return &dto.AuthResponse{
		Token: dto.TokenResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			TokenType:    tokenPair.TokenType,
			ExpiresIn:    tokenPair.ExpiresIn,
		},
		User: userEntity,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req dto.LogoutRequest) error {
	if req.AccessToken != "" {
		claims, err := s.tokenService.ValidateToken(req.AccessToken, TokenTypeAccess)
		if err == nil && claims != nil {
			_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeAccess, claims.ExpiresAt, "logout")
		}
	}

	if req.RefreshToken != "" {
		claims, err := s.tokenService.ValidateToken(req.RefreshToken, TokenTypeRefresh)
		if err == nil && claims != nil {
			_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeRefresh, claims.ExpiresAt, "logout")
		}
	}

	return nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, tokenStr string) (*TokenClaims, error) {
	claims, err := s.tokenService.ValidateToken(tokenStr, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

type APIKeyService struct {
	apiKeyRepo     APIKeyRepository
	keyGenerator   *APIKeyGenerator
	projectService *project.ProjectService
}

func NewAPIKeyService(apiKeyRepo APIKeyRepository, keyGenerator *APIKeyGenerator, projectService *project.ProjectService) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:     apiKeyRepo,
		keyGenerator: keyGenerator,
		projectService: projectService,
	}
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, projectID, userID uuid.UUID, req dto.CreateAPIKeyRequest) (*dto.CreatedAPIKeyResponse, error) {
	generated, err := s.keyGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure key: %w", err)
	}

	// Resolve the organization ID from the project to avoid storing uuid.Nil
	proj, err := s.projectService.GetProject(ctx, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	log.Printf("Project: %v", proj)

	apiKeyEntity := &APIKey{
		ID:             uuid.New(),
		OrganizationID: proj.OrganizationID,
		ProjectID:      projectID,
		UserID:         userID,
		Name:           req.Name,
		KeyPrefix:      generated.KeyPrefix,
		KeyHash:        generated.KeyHash,
		Status:         APIKeyStatusActive,
		ExpiresAt:      req.ExpiresAt,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.apiKeyRepo.Create(ctx, apiKeyEntity); err != nil {
		return nil, fmt.Errorf("failed to store API key: %w", err)
	}

	return &dto.CreatedAPIKeyResponse{
		ID:        apiKeyEntity.ID,
		Name:      apiKeyEntity.Name,
		SecretKey: generated.PlaintextKey,
		KeyPrefix: generated.KeyPrefix,
		ExpiresAt: apiKeyEntity.ExpiresAt,
		CreatedAt: apiKeyEntity.CreatedAt,
	}, nil
}

func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}

func (s *APIKeyService) ListProjectAPIKeys(ctx context.Context, projectID uuid.UUID) ([]*dto.APIKeyResponse, error) {
	keys, err := s.apiKeyRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.APIKeyResponse, len(keys))
	for i, k := range keys {
		responses[i] = &dto.APIKeyResponse{
			ID:             k.ID,
			OrganizationID: k.OrganizationID,
			ProjectID:      k.ProjectID,
			UserID:         k.UserID,
			Name:           k.Name,
			KeyPrefix:      k.KeyPrefix,
			Status:         string(k.Status),
			ExpiresAt:      k.ExpiresAt,
			LastUsedAt:     k.LastUsedAt,
			CreatedAt:      k.CreatedAt,
		}
	}
	return responses, nil
}

func (s *APIKeyService) AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*dto.AuthenticatedPrincipal, error) {
	if plaintextKey == "" {
		return nil, ErrAPIKeyNotFound
	}

	hash := HashAPIKey(plaintextKey)
	key, err := s.apiKeyRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	if !key.IsActive() {
		return nil, ErrAPIKeyInactive
	}

	_ = s.apiKeyRepo.UpdateLastUsed(ctx, key.ID)

	orgID := key.OrganizationID
	if orgID == uuid.Nil && s.projectService != nil {
		if proj, err := s.projectService.GetProject(ctx, key.ProjectID, key.UserID); err == nil && proj != nil {
			orgID = proj.OrganizationID
		}
	}

	return &dto.AuthenticatedPrincipal{
		APIKeyID:       key.ID,
		OrganizationID: orgID,
		ProjectID:      key.ProjectID,
		UserID:         key.UserID,
		KeyPrefix:      key.KeyPrefix,
	}, nil
}

func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID, currentJTI string, currentIP string, currentUA string) ([]SessionResponse, error) {
	if s.sessionRepo == nil {
		return []SessionResponse{}, nil
	}
	sessions, err := s.sessionRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []SessionResponse
	hasCurrent := false
	for _, sess := range sessions {
		isCurr := (currentJTI != "" && (sess.AccessJTI == currentJTI || sess.RefreshJTI == currentJTI))
		if isCurr {
			hasCurrent = true
		}
		result = append(result, SessionResponse{
			ID:           sess.ID,
			DeviceName:   sess.DeviceName,
			IPAddress:    sess.IPAddress,
			LastActiveAt: sess.LastActiveAt,
			CreatedAt:    sess.CreatedAt,
			IsCurrent:    isCurr,
		})
	}

	if !hasCurrent && currentJTI != "" {
		newSess := &UserSession{
			UserID:       userID,
			AccessJTI:    currentJTI,
			IPAddress:    currentIP,
			UserAgent:    currentUA,
			DeviceName:   ParseDeviceName(currentUA),
			LastActiveAt: time.Now().UTC(),
			ExpiresAt:    time.Now().UTC().Add(30 * 24 * time.Hour),
			IsRevoked:    false,
		}
		_ = s.sessionRepo.Create(ctx, newSess)
		result = append([]SessionResponse{{
			ID:           newSess.ID,
			DeviceName:   newSess.DeviceName,
			IPAddress:    newSess.IPAddress,
			LastActiveAt: newSess.LastActiveAt,
			CreatedAt:    newSess.CreatedAt,
			IsCurrent:    true,
		}}, result...)
	}

	return result, nil
}

func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentJTI string) error {
	if s.sessionRepo == nil {
		return nil
	}
	revoked, err := s.sessionRepo.RevokeOthers(ctx, userID, currentJTI)
	if err != nil {
		return err
	}

	if s.blacklistRepo != nil {
		for _, sess := range revoked {
			if sess.AccessJTI != "" {
				_ = s.blacklistRepo.BlacklistToken(ctx, sess.AccessJTI, userID, TokenTypeAccess, sess.ExpiresAt, "revoked_others")
			}
			if sess.RefreshJTI != "" {
				_ = s.blacklistRepo.BlacklistToken(ctx, sess.RefreshJTI, userID, TokenTypeRefresh, sess.ExpiresAt, "revoked_others")
			}
		}
	}
	return nil
}

func (s *AuthService) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	if s.sessionRepo == nil {
		return nil
	}
	sess, err := s.sessionRepo.RevokeByID(ctx, userID, sessionID)
	if err != nil {
		return err
	}
	if sess != nil && s.blacklistRepo != nil {
		if sess.AccessJTI != "" {
			_ = s.blacklistRepo.BlacklistToken(ctx, sess.AccessJTI, userID, TokenTypeAccess, sess.ExpiresAt, "revoked")
		}
		if sess.RefreshJTI != "" {
			_ = s.blacklistRepo.BlacklistToken(ctx, sess.RefreshJTI, userID, TokenTypeRefresh, sess.ExpiresAt, "revoked")
		}
	}
	return nil
}

