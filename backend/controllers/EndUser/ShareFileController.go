package EndUser

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"safesplit/models"
	"safesplit/services"
	"strconv"
	"strings"
	"time"

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
	ShareType models.ShareType `json:"share_type" binding:"required"`
	Password  string           `json:"password" binding:"required,min=6"`
	Email     string           `json:"email,omitempty"`
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
	// Get file ID from URL parameter
	fileID := ctx.Param("id")
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Convert string ID to uint
	id, err := strconv.ParseUint(fileID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if req.ShareType == models.RecipientShare && req.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email required for recipient share"})
		return
	}

	user := ctx.MustGet("user").(*models.User)
	file, err := c.fileModel.GetFileForDownload(uint(id), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	kek, err := services.DeriveKeyEncryptionKey(user.Password, user.MasterKeySalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	decryptedMasterKey, err := services.DecryptMasterKey(user.EncryptedMasterKey, kek, user.MasterKeyNonce)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Decryption failed"})
		return
	}

	userMasterKey := decryptedMasterKey[:32]
	fragments, err := c.keyFragmentModel.GetUserFragmentsForFile(file.ID)
	if err != nil || len(fragments) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fragments"})
		return
	}

	userFragment := fragments[0]
	decryptedFragment, err := services.DecryptMasterKey(
		userFragment.Data,
		userMasterKey,
		userFragment.EncryptionNonce,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Fragment decryption failed"})
		return
	}

	encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
		decryptedFragment,
		[]byte(req.Password),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Fragment encryption failed"})
		return
	}

	share := &models.FileShare{
		FileID:               file.ID,
		SharedBy:             user.ID,
		EncryptedKeyFragment: encryptedFragment,
		FragmentIndex:        userFragment.FragmentIndex,
		IsActive:             true,
		ShareType:            req.ShareType,
		Email:                req.Email,
	}

	if err := c.fileShareModel.CreateFileShare(share, req.Password); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Share creation failed"})
		return
	}

	// Get base URL from environment variable
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Create the complete share URL
	shareURL := fmt.Sprintf("%s/api/files/share/%s", baseURL, share.ShareLink)

	if req.ShareType == models.RecipientShare {
		// Get base URL from environment variable
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}

		// Create the frontend share URL (not the API URL)
		shareURL := fmt.Sprintf("%s/protected-share/%s", baseURL, share.ShareLink)

		emailBody := fmt.Sprintf(`Hello,
	
	You have received a secure file share from %s.
	
	File: %s
	Access Link: %s
	
	This link requires a password and email verification to access. 
	Please use the same email address this message was sent to when accessing the file.
	
	Best regards,
	SafeSplit Team`, user.Username, file.OriginalName, shareURL)

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
		Details:      fmt.Sprintf("Created %s share", req.ShareType),
	})

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

	// For GET requests, return file info and requirements
	if ctx.Request.Method == "GET" {
		share, err := c.fileShareModel.GetShareByLink(shareLink)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "Invalid share"})
			return
		}

		// Get file info
		file, err := c.fileModel.GetFileByID(share.FileID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "File not found"})
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

	// Handle POST request
	var req AccessShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request"})
		return
	}

	share, err := c.fileShareModel.GetShareByLink(shareLink)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid share"})
		return
	}

	// Check if share has expired
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Share link has expired"})
		return
	}

	// Check if maximum downloads reached
	if share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Maximum number of downloads reached"})
		return
	}

	// Check if share is still active
	if !share.IsActive {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Share link is no longer active"})
		return
	}

	// Validate share based on type
	if share.ShareType == models.RecipientShare {
		// Validate password only
		share, validationErr := c.fileShareModel.ValidateRecipientShare(shareLink, req.Password)
		if validationErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "Invalid password"})
			return
		}

		// Send 2FA to the email associated with the share
		if err := c.twoFactorService.SendTwoFactorToken(share.ID, share.Email); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to send 2FA code"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "2FA code sent to registered email",
			"data": gin.H{
				"share_id": share.ID,
			},
		})
		return
	} else {
		// For normal shares, just validate password
		share, err := c.fileShareModel.ValidateShare(shareLink, req.Password)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "Invalid password"})
			return
		}
		c.processFileAccess(ctx, share, req.Password)
	}
}
func (c *ShareFileController) Verify2FAAndDownload(ctx *gin.Context) {
	shareLink := ctx.Param("shareLink")
	var req TwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data"})
		return
	}

	share, err := c.fileShareModel.GetShareByLink(shareLink)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid share"})
		return
	}

	// Verify share type
	if share.ShareType != models.RecipientShare {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "2FA verification only required for recipient shares"})
		return
	}

	// Verify 2FA code
	if err := c.twoFactorService.VerifyToken(share.ID, req.Code); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid 2FA code"})
		return
	}

	// Process file download
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server key data error"})
		return
	}

	sharedDecryptedFragment, err := c.encryptionService.DecryptKeyFragment(
		share.EncryptedKeyFragment,
		[]byte(password),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Fragment decryption failed"})
		return
	}

	shares := make([]services.KeyShare, file.Threshold)
	usedIndices := make(map[int]bool)

	shares[0] = services.KeyShare{
		Index: share.FragmentIndex,
		Value: hex.EncodeToString(sharedDecryptedFragment),
	}
	usedIndices[share.FragmentIndex] = true

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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient unique shares"})
		return
	}

	var encryptedData []byte
	var retrievalErr error

	if file.IsSharded {
		encryptedData, retrievalErr = c.getShardedData(file)
	} else {
		encryptedData, retrievalErr = c.fileModel.ReadFileContent(file.FilePath)
	}

	if retrievalErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "File retrieval failed"})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "File decryption failed"})
		return
	}

	// Handle decompression if the file is compressed
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

	if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
		return nil, fmt.Errorf("insufficient shards for reconstruction")
	}

	return c.rsService.ReconstructFile(fileShards.Shards,
		int(file.DataShardCount), int(file.ParityShardCount))
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
