package PremiumUser

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"safesplit/models"
	"safesplit/services"

	"github.com/gin-gonic/gin"
)

type ShareFileController struct {
	fileModel         *models.FileModel
	fileShareModel    *models.FileShareModel
	keyFragmentModel  *models.KeyFragmentModel
	encryptionService *services.EncryptionService
	activityLogModel  *models.ActivityLogModel
	rsService         *services.ReedSolomonService
	userModel         *models.UserModel
	serverKeyModel    *models.ServerMasterKeyModel
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
) *ShareFileController {
	return &ShareFileController{
		fileModel:         fileModel,
		fileShareModel:    fileShareModel,
		keyFragmentModel:  keyFragmentModel,
		encryptionService: encryptionService,
		activityLogModel:  activityLogModel,
		rsService:         rsService,
		userModel:         userModel,
		serverKeyModel:    serverKeyModel,
	}
}

type CreateShareRequest struct {
	FileID       uint       `json:"file_id" binding:"required"`
	Password     string     `json:"password" binding:"required,min=6"`
	ExpiresAt    *time.Time `json:"expires_at"`
	MaxDownloads *int       `json:"max_downloads"`
}

type AccessShareRequest struct {
	Password string `json:"password" binding:"required"`
}

func (c *ShareFileController) CreateShare(ctx *gin.Context) {
	var req CreateShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid request data",
		})
		return
	}

	// Validate expiry date if provided
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Expiry date cannot be in the past",
		})
		return
	}

	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return
	}
	currentUser := user.(*models.User)

	// Get file and verify ownership
	file, err := c.fileModel.GetFileForDownload(req.FileID, currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found or access denied",
		})
		return
	}

	// Derive user's KEK
	kek, err := services.DeriveKeyEncryptionKey(currentUser.Password, currentUser.MasterKeySalt)
	if err != nil {
		log.Printf("Failed to derive KEK: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process encryption",
		})
		return
	}

	// Decrypt user's master key
	decryptedMasterKey, err := services.DecryptMasterKey(
		currentUser.EncryptedMasterKey,
		kek,
		currentUser.MasterKeyNonce,
	)
	if err != nil {
		log.Printf("Failed to decrypt master key: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process encryption",
		})
		return
	}

	// Use first 32 bytes of decrypted master key
	userMasterKey := decryptedMasterKey[:32]

	// Get fragments
	fragments, err := c.keyFragmentModel.GetUserFragmentsForFile(file.ID)
	if err != nil || len(fragments) == 0 {
		log.Printf("Failed to retrieve key fragments for file %d: %v", file.ID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return
	}

	log.Printf("Creating premium share for file %d with %d fragments", file.ID, len(fragments))
	userFragment := fragments[0]
	log.Printf("Selected user fragment with index %d for sharing", userFragment.FragmentIndex)

	// Decrypt the fragment using master key
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

	// Encrypt fragment with share password
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

	// Create premium share record
	share := &models.FileShare{
		FileID:               file.ID,
		SharedBy:             currentUser.ID,
		EncryptedKeyFragment: encryptedFragment,
		FragmentIndex:        userFragment.FragmentIndex,
		ExpiresAt:            req.ExpiresAt,
		MaxDownloads:         req.MaxDownloads,
		IsActive:             true,
	}

	if err := c.fileShareModel.CreateFileShare(share, req.Password); err != nil {
		log.Printf("Failed to create file share: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create share",
		})
		return
	}

	// Log premium share creation
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       currentUser.ID,
		ActivityType: "premium_share",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("Premium share created (Expires: %v, Max Downloads: %v)", req.ExpiresAt, req.MaxDownloads),
	}); err != nil {
		log.Printf("Failed to log share activity: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"share_link": share.ShareLink,
		},
	})
}

func (c *ShareFileController) AccessShare(ctx *gin.Context) {
	shareLink := ctx.Param("shareLink")
	var req AccessShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid password",
		})
		return
	}

	// Validate share with premium features
	share, err := c.fileShareModel.ValidateShare(shareLink, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid share link or password",
		})
		return
	}

	// Check if share has expired
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Share link has expired",
		})
		return
	}

	// Check if maximum downloads reached
	if share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Maximum number of downloads reached",
		})
		return
	}

	// Check if share is still active
	if !share.IsActive {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "Share link is no longer active",
		})
		return
	}

	log.Printf("Processing premium share access for link: %s", shareLink)

	// Get file metadata
	file, err := c.fileModel.GetFileByID(share.FileID)
	if err != nil {
		log.Printf("Failed to get file %d: %v", share.FileID, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "File not found",
		})
		return
	}

	// Get server fragments
	serverFragments, err := c.keyFragmentModel.GetServerFragmentsForFile(share.FileID)
	if err != nil {
		log.Printf("Failed to get server fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file access",
		})
		return
	}

	log.Printf("Retrieved %d server fragments", len(serverFragments))

	// Verify we have enough fragments
	if len(serverFragments)+1 < int(file.Threshold) {
		log.Printf("Insufficient fragments: have %d server + 1 shared, need %d",
			len(serverFragments), file.Threshold)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Insufficient fragments to reconstruct file",
		})
		return
	}

	// Get server key for decrypting server fragments
	serverKey, err := c.serverKeyModel.GetActive()
	if err != nil {
		log.Printf("Failed to get server key: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process decryption",
		})
		return
	}

	serverKeyData, err := c.serverKeyModel.GetServerKey(serverKey.KeyID)
	if err != nil {
		log.Printf("Failed to get server key data: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to get server key",
		})
		return
	}

	// Decrypt shared fragment
	sharedDecryptedFragment, err := c.encryptionService.DecryptKeyFragment(
		share.EncryptedKeyFragment,
		[]byte(req.Password),
	)
	if err != nil {
		log.Printf("Failed to decrypt shared fragment: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process file decryption",
		})
		return
	}

	// Prepare key shares array
	shares := make([]services.KeyShare, file.Threshold)
	usedIndices := make(map[int]bool)

	// Add shared fragment first
	shares[0] = services.KeyShare{
		Index: share.FragmentIndex,
		Value: hex.EncodeToString(sharedDecryptedFragment),
	}
	usedIndices[share.FragmentIndex] = true
	log.Printf("Added shared fragment with index %d", share.FragmentIndex)

	// Add server fragments
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
		log.Printf("Added server fragment %d with index %d", i, fragment.FragmentIndex)
		sharesAdded++
	}

	if sharesAdded < file.Threshold {
		log.Printf("Failed to get enough unique shares: have %d, need %d", sharesAdded, file.Threshold)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to get enough unique shares",
		})
		return
	}

	// Get encrypted file data
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read file data",
		})
		return
	}

	// Decrypt the file with encryption type
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to decrypt file",
		})
		return
	}

	// Update premium share status
	if err := c.fileShareModel.IncrementDownloadCount(share.ID); err != nil {
		log.Printf("Failed to increment download count: %v", err)
	}

	// Log premium share access
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       share.SharedBy,
		ActivityType: "premium_download",
		FileID:       &file.ID,
		IPAddress:    ctx.ClientIP(),
		Status:       "success",
		Details:      fmt.Sprintf("Premium shared file download using %d fragments", file.Threshold),
	}); err != nil {
		log.Printf("Failed to log share download activity: %v", err)
	}

	c.sendFileResponse(ctx, file, decryptedData)
}

func (c *ShareFileController) getShardedData(file *models.File) ([]byte, error) {
	fileShards, err := c.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shards: %w", err)
	}

	// Log shard information for debugging
	validShards := 0
	for i, shard := range fileShards.Shards {
		if shard != nil {
			validShards++
			log.Printf("Shard %d: %d bytes", i, len(shard))
		} else {
			log.Printf("Shard %d: Missing", i)
		}
	}

	// Validate we have enough shards for reconstruction
	if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
		return nil, fmt.Errorf("insufficient shards for reconstruction: have %d, need %d",
			validShards, file.DataShardCount)
	}

	// Reconstruct file from shards
	reconstructed, err := c.rsService.ReconstructFile(fileShards.Shards,
		int(file.DataShardCount), int(file.ParityShardCount))
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct file: %w", err)
	}

	log.Printf("Successfully reconstructed file data: %d bytes", len(reconstructed))
	return reconstructed, nil
}

func (c *ShareFileController) sendFileResponse(ctx *gin.Context, file *models.File, data []byte) {
	sanitizedFilename := strings.ReplaceAll(file.OriginalName, `"`, `\"`)
	encodedFilename := url.QueryEscape(sanitizedFilename)

	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, Content-Length")
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
		sanitizedFilename, encodedFilename))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	ctx.Header("X-Original-Filename", url.QueryEscape(file.OriginalName))

	log.Printf("Sending file response: %s (Size: %d bytes)", file.OriginalName, len(data))
	ctx.Data(http.StatusOK, file.MimeType, data)
}
