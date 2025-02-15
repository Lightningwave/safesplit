package PremiumUser

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
)

type ShareFileController struct {
	fileModel          *models.FileModel
	fileShareModel     *models.FileShareModel
	keyFragmentModel   *models.KeyFragmentModel
	encryptionService  *services.EncryptionService
	activityLogModel   *models.ActivityLogModel
	rsService          *services.ReedSolomonService
	userModel          *models.UserModel
	serverKeyModel     *models.ServerMasterKeyModel
	twoFactorService   *services.TwoFactorAuthService
	emailService       *services.SMTPEmailService
	compressionService *services.CompressionService
}

func NewShareFileController(
	fileModel *models.FileModel,
	fileShareModel *models.FileShareModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	activityLogModel *models.ActivityLogModel,
	rsService *services.ReedSolomonService,
	userModel *models.UserModel,
	serverKeyModel *models.ServerMasterKeyModel,
	twoFactorService *services.TwoFactorAuthService,
	emailService *services.SMTPEmailService,
	compressionService *services.CompressionService,
) *ShareFileController {
	return &ShareFileController{
		fileModel:          fileModel,
		fileShareModel:     fileShareModel,
		keyFragmentModel:   keyFragmentModel,
		encryptionService:  encryptionService,
		activityLogModel:   activityLogModel,
		rsService:          rsService,
		userModel:          userModel,
		serverKeyModel:     serverKeyModel,
		twoFactorService:   twoFactorService,
		emailService:       emailService,
		compressionService: compressionService,
	}
}

type CreateShareRequest struct {
	Password     string           `json:"password" binding:"required,min=6"`
	ExpiresAt    *time.Time       `json:"expires_at"`
	MaxDownloads *int             `json:"max_downloads"`
	ShareType    models.ShareType `json:"share_type" binding:"required"`
	Email        string           `json:"email,omitempty"`
}

type AccessShareRequest struct {
	Password string `json:"password" binding:"required"`
	Email    string `json:"email,omitempty"`
}

type TwoFactorRequest struct {
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *ShareFileController) CreateShare(ctx *gin.Context) {
	log.Printf("Received advanced share creation request for file ID: %v", ctx.Param("id"))

	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	if req.ShareType == models.RecipientShare && req.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Email required for recipient share",
		})
		return
	}

	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Expiry date cannot be in the past",
		})
		return
	}

	user := ctx.MustGet("user").(*models.User)

	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.fileModel.GetFileForDownload(uint(fileID), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found or access denied",
		})
		return
	}

	kek, err := services.DeriveKeyEncryptionKey(user.Password, user.MasterKeySalt)
	if err != nil {
		log.Printf("Failed to derive KEK: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process encryption",
		})
		return
	}

	decryptedMasterKey, err := services.DecryptMasterKey(
		user.EncryptedMasterKey,
		kek,
		user.MasterKeyNonce,
	)
	if err != nil {
		log.Printf("Failed to decrypt master key: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process encryption",
		})
		return
	}

	userMasterKey := decryptedMasterKey[:32]

	fragments, err := c.keyFragmentModel.GetUserFragmentsForFile(file.ID)
	if err != nil || len(fragments) == 0 {
		log.Printf("Failed to retrieve key fragments for file %d: %v", file.ID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return
	}

	userFragment := fragments[0]
	decryptedFragment, err := services.DecryptMasterKey(
		userFragment.Data,
		userMasterKey,
		userFragment.EncryptionNonce,
	)
	if err != nil {
		log.Printf("Failed to decrypt fragment: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process share creation",
		})
		return
	}

	encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
		decryptedFragment,
		[]byte(req.Password),
	)
	if err != nil {
		log.Printf("Failed to encrypt key fragment: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process share encryption",
		})
		return
	}

	share := &models.FileShare{
		FileID:               file.ID,
		SharedBy:             user.ID,
		EncryptedKeyFragment: encryptedFragment,
		FragmentIndex:        userFragment.FragmentIndex,
		ExpiresAt:            req.ExpiresAt,
		MaxDownloads:         req.MaxDownloads,
		IsActive:             true,
		ShareType:            req.ShareType,
		Email:                req.Email,
	}

	if err := c.fileShareModel.CreateFileShare(share, req.Password); err != nil {
		log.Printf("Failed to create file share: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create share",
		})
		return
	}

	if req.ShareType == models.RecipientShare {
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "https://safesplit.xyz"
		}

		shareURL := fmt.Sprintf("%s/protected-share/%s", baseURL, share.ShareLink)

		emailBody := fmt.Sprintf(`Hello,

You have received a secure file share from %s.

File: %s
Access Link: %s
Email: %s
Password: %s

This link requires password and email verification to access. 
Please use the same email address this message was sent to when accessing the file.

Best regards,
SafeSplit Team`, user.Username, file.OriginalName, shareURL, req.Email, req.Password)

		if err := c.emailService.SendEmail(
			req.Email,
			"Secure File Share Received",
			emailBody,
		); err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	}

	c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       user.ID,
		ActivityType: "share",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("Created %s share with premium features", req.ShareType),
	})

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://safesplit.xyz"
	}

	sharePath := "/premium/share/"
	if req.ShareType == models.RecipientShare {
		sharePath = "/protected-share/"
	}

	shareURL := fmt.Sprintf("%s%s%s", baseURL, sharePath, share.ShareLink)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"share_link":   shareURL,
			"raw_link":     share.ShareLink,
			"requires_2fa": req.ShareType == models.RecipientShare,
		},
	})
}

func (c *ShareFileController) AccessShare(ctx *gin.Context) {
	shareLink := ctx.Param("shareLink")
	log.Printf("Received share access request for link: %s", shareLink)

	if ctx.Request.Method == "GET" {
		share, err := c.fileShareModel.GetShareByLink(shareLink)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid share"})
			return
		}

		file, err := c.fileModel.GetFileByID(share.FileID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"requires_password": true,
				"requires_2fa":      share.ShareType == models.RecipientShare,
				"recipient_share":   share.ShareType == models.RecipientShare,
				"file_name":         file.OriginalName,
				"file_size":         file.Size,
				"mime_type":         file.MimeType,
				"created_at":        share.CreatedAt,
				"expires_at":        share.ExpiresAt,
				"download_count":    share.DownloadCount,
				"max_downloads":     share.MaxDownloads,
			},
		})
		return
	}

	var req AccessShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	share, err := c.fileShareModel.GetShareByLink(shareLink)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid share"})
		return
	}

	// Get file info early for use in verification
	file, err := c.fileModel.GetFileByID(share.FileID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if share.ShareType == models.RecipientShare {
		share, validationErr := c.fileShareModel.ValidateRecipientShare(shareLink, req.Password)
		if validationErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "Invalid password"})
			return
		}

		// Send verification code to the email associated with the share
		if err := c.twoFactorService.SendShareVerificationToken(share.ID, share.Email, file.OriginalName); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to send verification code"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Verification code sent to registered email",
			"data": gin.H{
				"share_id": share.ID,
			},
		})
		return
	}

	if err := c.validatePremiumShare(share); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.processFileAccess(ctx, share, req.Password)
}
func (c *ShareFileController) validatePremiumShare(share *models.FileShare) error {
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return fmt.Errorf("share link has expired")
	}

	if share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads {
		return fmt.Errorf("maximum number of downloads reached")
	}

	if !share.IsActive {
		return fmt.Errorf("share link is no longer active")
	}

	return nil
}

func (c *ShareFileController) Verify2FAAndDownload(ctx *gin.Context) {
	shareLink := ctx.Param("shareLink")
	var req TwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	share, err := c.fileShareModel.GetShareByLink(shareLink)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid share"})
		return
	}

	if share.ShareType != models.RecipientShare {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "2FA verification only required for recipient shares"})
		return
	}

	if err := c.twoFactorService.VerifyToken(share.ID, req.Code); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA code"})
		return
	}

	if err := c.validatePremiumShare(share); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.processFileAccess(ctx, share, req.Password)
}

func (c *ShareFileController) processFileAccess(ctx *gin.Context, share *models.FileShare, password string) {
	file, err := c.fileModel.GetFileByID(share.FileID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	serverFragments, err := c.keyFragmentModel.GetServerFragmentsForFile(share.FileID)
	if err != nil || len(serverFragments)+1 < int(file.Threshold) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient fragments"})
		return
	}

	serverKey, err := c.serverKeyModel.GetActive()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server key error"})
		return
	}

	serverKeyData, err := c.serverKeyModel.GetServerKey(serverKey.KeyID)
	if err != nil {
		log.Printf("Failed to get server key data: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get server key"})
		return
	}

	sharedDecryptedFragment, err := c.encryptionService.DecryptKeyFragment(
		share.EncryptedKeyFragment,
		[]byte(password),
	)
	if err != nil {
		log.Printf("Failed to decrypt shared fragment: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file decryption"})
		return
	}

	shares := make([]services.KeyShare, file.Threshold)
	usedIndices := make(map[int]bool)

	shares[0] = services.KeyShare{
		Index: share.FragmentIndex,
		Value: hex.EncodeToString(sharedDecryptedFragment),
	}
	usedIndices[share.FragmentIndex] = true
	log.Printf("Added shared fragment with index %d", share.FragmentIndex)

	sharesAdded := uint(1)
	for i := 0; i < len(serverFragments) && sharesAdded < file.Threshold; i++ {
		fragment := serverFragments[i]
		if usedIndices[fragment.FragmentIndex] {
			continue
		}

		decryptedFragment, err := services.DecryptMasterKey(
			fragment.Data,
			serverKeyData,
			fragment.EncryptionNonce,
		)
		if err != nil {
			log.Printf("Failed to decrypt server fragment %d: %v", i, err)
			continue
		}

		shares[sharesAdded] = services.KeyShare{
			Index:        fragment.FragmentIndex,
			Value:        hex.EncodeToString(decryptedFragment),
			NodeIndex:    fragment.NodeIndex,
			FragmentPath: fragment.FragmentPath,
		}
		usedIndices[fragment.FragmentIndex] = true
		sharesAdded++
	}

	if sharesAdded < file.Threshold {
		log.Printf("Failed to get enough unique shares: have %d, need %d", sharesAdded, file.Threshold)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get enough unique shares"})
		return
	}

	var encryptedData []byte
	var retrievalErr error

	if file.IsSharded {
		log.Printf("Retrieving sharded data for file %d", file.ID)
		encryptedData, retrievalErr = c.getShardedData(file)
	} else {
		log.Printf("Reading file content from path: %s", file.FilePath)
		encryptedData, retrievalErr = c.fileModel.ReadFileContent(file.FilePath)
	}

	if retrievalErr != nil {
		log.Printf("Failed to retrieve file data: %v", retrievalErr)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file data"})
		return
	}

	decryptedData, err := c.encryptionService.DecryptFileWithType(
		encryptedData,
		file.EncryptionIV,
		shares,
		int(file.Threshold),
		file.EncryptionSalt,
		file.EncryptionType,
	)
	if err != nil {
		log.Printf("Failed to decrypt file data: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt file"})
		return
	}

	if file.IsCompressed {
		log.Printf("Decompressing data for file ID: %d", file.ID)
		decryptedData, err = c.compressionService.Decompress(decryptedData)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decompress file"})
			return
		}
	}

	if err := c.fileShareModel.IncrementDownloadCount(share.ID); err != nil {
		log.Printf("Failed to increment download count: %v", err)
	}

	c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       share.SharedBy,
		ActivityType: "download",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("Download with %d fragments", file.Threshold),
	})

	c.sendFileResponse(ctx, file, decryptedData)
}

func (c *ShareFileController) getShardedData(file *models.File) ([]byte, error) {
	fileShards, err := c.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shards: %w", err)
	}

	validShards := 0
	for i, shard := range fileShards.Shards {
		if shard != nil {
			validShards++
			log.Printf("Shard %d: %d bytes", i, len(shard))
		} else {
			log.Printf("Shard %d: Missing", i)
		}
	}

	if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
		return nil, fmt.Errorf("insufficient shards for reconstruction: have %d, need %d",
			validShards, file.DataShardCount)
	}

	reconstructed, err := c.rsService.ReconstructFile(fileShards.Shards,
		int(file.DataShardCount), int(file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct file: %w", err)
	}

	log.Printf("Successfully reconstructed file data: %d bytes", len(reconstructed))
	return reconstructed, nil
}

func (c *ShareFileController) sendFileResponse(ctx *gin.Context, file *models.File, data []byte) {
	escapedName := strings.ReplaceAll(file.OriginalName, `"`, `\"`)
	utf8Name := url.PathEscape(file.OriginalName)
	ctx.Header("Content-Disposition", fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		escapedName,
		utf8Name,
	))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	ctx.Header("X-Original-Filename", escapedName)
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, Content-Length, X-Original-Filename")
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	log.Printf("Sending file response: %s (Size: %d bytes)", file.OriginalName, len(data))
	ctx.Data(http.StatusOK, file.MimeType, data)
}
