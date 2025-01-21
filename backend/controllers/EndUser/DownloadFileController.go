package EndUser

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"safesplit/models"
	"safesplit/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DownloadFileController struct {
	fileModel          *models.FileModel
	keyFragmentModel   *models.KeyFragmentModel
	encryptionService  *services.EncryptionService
	activityLogModel   *models.ActivityLogModel
	compressionService *services.CompressionService
	rsService          *services.ReedSolomonService
	serverKeyModel     *models.ServerMasterKeyModel
}

func NewDownloadFileController(
	fileModel *models.FileModel,
	keyFragmentModel *models.KeyFragmentModel,
	encryptionService *services.EncryptionService,
	activityLogModel *models.ActivityLogModel,
	compressionService *services.CompressionService,
	rsService *services.ReedSolomonService,
	serverKeyModel *models.ServerMasterKeyModel, 
) *DownloadFileController {
	return &DownloadFileController{
		fileModel:          fileModel,
		keyFragmentModel:   keyFragmentModel,
		encryptionService:  encryptionService,
		activityLogModel:   activityLogModel,
		compressionService: compressionService,
		rsService:          rsService,
		serverKeyModel:     serverKeyModel, 
	}
}

func (c *DownloadFileController) Download(ctx *gin.Context) {
	log.Printf("Starting file download request")

	currentUser, err := c.getCurrentUser(ctx)
	if err != nil {
		return
	}

	fileID, err := c.getFileID(ctx)
	if err != nil {
		return
	}

	file, err := c.getAndValidateFile(ctx, currentUser.ID, fileID)
	if err != nil {
		return
	}

	log.Printf("Processing file ID: %d, IsSharded: %v, IsCompressed: %v",
		file.ID, file.IsSharded, file.IsCompressed)

	// Get key shares first to fail early if they're not available
	shares, err := c.getKeyShares(ctx, file)
	if err != nil {
		log.Printf("Failed to get key shares: %v", err)
		return
	}

	// Get file data based on storage type
	encryptedData, err := c.getFileData(ctx, file)
	if err != nil {
		return
	}

	// Decrypt the data
	decryptedData, err := c.decryptData(ctx, file, encryptedData, shares)
	if err != nil {
		return
	}

	// Decompress if needed
	finalData, err := c.handleDecompression(ctx, file, decryptedData)
	if err != nil {
		return
	}

	// Log success and send response
	c.logDownloadActivity(currentUser, file, ctx.ClientIP())
	c.sendFileResponse(ctx, file, finalData)
}

func (c *DownloadFileController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
	user, exists := ctx.Get("user")
	if !exists {
		log.Printf("User authentication failed - user not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Unauthorized access",
		})
		return nil, fmt.Errorf("unauthorized")
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		log.Printf("User authentication failed - invalid user type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Invalid user data",
		})
		return nil, fmt.Errorf("invalid user data")
	}

	return currentUser, nil
}

func (c *DownloadFileController) getFileID(ctx *gin.Context) (uint, error) {
	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		log.Printf("Invalid file ID: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid file ID",
		})
		return 0, err
	}
	return uint(fileID), nil
}

func (c *DownloadFileController) getAndValidateFile(ctx *gin.Context, userID, fileID uint) (*models.File, error) {
	file, err := c.fileModel.GetFileForDownload(fileID, userID)
	if err != nil {
		log.Printf("Error fetching file: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "file not found or access denied" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return nil, err
	}

	if err := c.validateFileMetadata(file); err != nil {
		log.Printf("File validation failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return nil, err
	}

	log.Printf("Retrieved file: ID=%d, IsSharded=%v, Path=%s, Salt length=%d",
		file.ID, file.IsSharded, file.FilePath, len(file.EncryptionSalt))
	return file, nil
}

func (c *DownloadFileController) validateFileMetadata(file *models.File) error {
	var expectedIVSize int
	switch file.EncryptionType {
	case services.ChaCha20:
		expectedIVSize = 24 // XChaCha20-Poly1305
	case services.Twofish:
		expectedIVSize = 12 // GCM requires 12-byte nonce for Twofish
	case services.StandardEncryption:
		expectedIVSize = 16 // AES-GCM
	default:
		return fmt.Errorf("unsupported encryption type: %s", file.EncryptionType)
	}

	if len(file.EncryptionIV) != expectedIVSize {
		return fmt.Errorf("invalid IV length for %s encryption: got %d, expected %d",
			file.EncryptionType, len(file.EncryptionIV), expectedIVSize)
	}

	if len(file.EncryptionSalt) != 32 {
		return fmt.Errorf("invalid salt length: got %d, expected 32", len(file.EncryptionSalt))
	}

	if file.Threshold < 2 {
		return fmt.Errorf("invalid threshold value: %d", file.Threshold)
	}

	if file.IsSharded {
		if file.DataShardCount < 1 || file.ParityShardCount < 1 {
			return fmt.Errorf("invalid shard configuration: data=%d, parity=%d",
				file.DataShardCount, file.ParityShardCount)
		}
	} else {
		if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found on server at path: %s", file.FilePath)
		}
	}

	return nil
}

func (c *DownloadFileController) getKeyShares(ctx *gin.Context, file *models.File) ([]services.KeyShare, error) {
	log.Printf("Retrieving key fragments for file %d - Threshold: %d, Expected shares: %d",
		file.ID, file.Threshold, file.ShareCount)

	// Get the server master key
	serverKey, err := c.serverKeyModel.GetActive()
	if err != nil {
		log.Printf("Failed to retrieve server key: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve server key",
		})
		return nil, err
	}

	// Get the decrypted server key data
	serverKeyData, err := c.serverKeyModel.GetServerKey(serverKey.KeyID)
	if err != nil {
		log.Printf("Failed to get server key data: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to get server key data",
		})
		return nil, err
	}

	// Get user
	currentUser, err := c.getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get fragments with their data
	fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
	if err != nil {
		log.Printf("Failed to retrieve key fragments: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve key fragments",
		})
		return nil, err
	}

	log.Printf("Retrieved %d total fragments for file %d", len(fragments), file.ID)

	shares := make([]services.KeyShare, len(fragments))
	for i, fragment := range fragments {
		var decryptedFragment []byte

		// Decrypt fragment based on its holder type
		if fragment.KeyFragment.HolderType == models.ServerHolder {
			log.Printf("Decrypting server fragment %d with server key", i)
			decryptedFragment, err = services.DecryptMasterKey(
				fragment.Data,
				serverKeyData,
				fragment.KeyFragment.EncryptionNonce[:12],
			)
		} else {
			log.Printf("Decrypting user fragment %d with user master key", i)
			userMasterKey := currentUser.EncryptedMasterKey[:32]
			decryptedFragment, err = services.DecryptMasterKey(
				fragment.Data,
				userMasterKey,
				fragment.KeyFragment.EncryptionNonce[:12],
			)
		}

		if err != nil {
			log.Printf("Failed to decrypt fragment %d: %v", i, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  fmt.Sprintf("Failed to decrypt fragment: %v", err),
			})
			return nil, err
		}

		// Normalize to 32 bytes
		normalizedFragment := make([]byte, 32)
		copy(normalizedFragment, decryptedFragment)

		log.Printf("Fragment %d raw decrypted length: %d", i, len(decryptedFragment))
		log.Printf("Fragment %d raw bytes: %x", i, decryptedFragment)

		// Convert to hex string to match encryption format
		shares[i] = services.KeyShare{
			Index:        fragment.KeyFragment.FragmentIndex,
			Value:        hex.EncodeToString(normalizedFragment),
			NodeIndex:    fragment.KeyFragment.NodeIndex,
			FragmentPath: fragment.KeyFragment.FragmentPath,
		}

		log.Printf("Fragment %d normalized: Index=%d, Node=%d, Path=%s, Length=%d, Bytes=%x, HexValue=%s",
			i, shares[i].Index, shares[i].NodeIndex, shares[i].FragmentPath,
			len(normalizedFragment), normalizedFragment, shares[i].Value)
	}

	return shares, nil
}
func (c *DownloadFileController) getFileData(ctx *gin.Context, file *models.File) ([]byte, error) {
	log.Printf("Getting file data for file ID: %d", file.ID)

	if file.IsSharded {
		// Handle sharded files using Reed-Solomon
		return c.getShardedData(ctx, file)
	}

	// Handle regular files
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		log.Printf("Failed to read file from disk: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to read file data",
		})
		return nil, err
	}

	log.Printf("Successfully read file data - Size: %d bytes", len(data))
	return data, nil
}

func (c *DownloadFileController) getShardedData(ctx *gin.Context, file *models.File) ([]byte, error) {
	log.Printf("Beginning shard retrieval for file %d - Data shards: %d, Parity shards: %d",
		file.ID, file.DataShardCount, file.ParityShardCount)

	fileShards, err := c.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
	if err != nil {
		log.Printf("Failed to retrieve shards: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve file shards",
		})
		return nil, err
	}

	validShards := 0
	for i, shard := range fileShards.Shards {
		if shard != nil {
			validShards++
			log.Printf("Shard %d: Length = %d bytes", i, len(shard))
		} else {
			log.Printf("Shard %d: Missing", i)
		}
	}
	log.Printf("Retrieved %d valid shards out of %d total shards",
		validShards, file.DataShardCount+file.ParityShardCount)

	if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
		log.Printf("Insufficient shards for file %d - Need %d data shards, have %d valid shards",
			file.ID, file.DataShardCount, validShards)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Insufficient shards available for reconstruction",
		})
		return nil, fmt.Errorf("insufficient shards")
	}

	reconstructed, err := c.rsService.ReconstructFile(fileShards.Shards, int(file.DataShardCount), int(file.ParityShardCount))
	if err != nil {
		log.Printf("Failed to reconstruct file from shards: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to reconstruct file",
		})
		return nil, err
	}

	log.Printf("Successfully reconstructed file from shards - Total size: %d bytes", len(reconstructed))
	return reconstructed, nil
}

func (c *DownloadFileController) decryptData(ctx *gin.Context, file *models.File, data []byte, shares []services.KeyShare) ([]byte, error) {
	log.Printf("Starting decryption for file ID: %d with encryption type: %s", file.ID, file.EncryptionType)
	log.Printf("Data size: %d bytes", len(data))
	log.Printf("IV length: %d bytes", len(file.EncryptionIV))
	log.Printf("Salt length: %d bytes", len(file.EncryptionSalt))

	for i, share := range shares {
		log.Printf("Share %d - Index: %d, Value length: %d", i, share.Index, len(share.Value))
	}

	// Use the appropriate decryption method based on encryption type
	decrypted, err := c.encryptionService.DecryptFileWithType(
		data,
		file.EncryptionIV,
		shares,
		int(file.Threshold),
		file.EncryptionSalt,
		file.EncryptionType,
	)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Failed to decrypt file: %v", err),
		})
		return nil, err
	}

	log.Printf("Successfully decrypted data - Decrypted size: %d bytes", len(decrypted))
	return decrypted, nil
}

func (c *DownloadFileController) handleDecompression(ctx *gin.Context, file *models.File, data []byte) ([]byte, error) {
	if !file.IsCompressed {
		return data, nil
	}

	log.Printf("Decompressing data for file ID: %d", file.ID)
	decompressed, err := c.compressionService.Decompress(data)
	if err != nil {
		log.Printf("Decompression failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to decompress file",
		})
		return nil, err
	}
	return decompressed, nil
}

func (c *DownloadFileController) logDownloadActivity(user *models.User, file *models.File, ipAddress string) {
	activityDetail := "File downloaded successfully"
	if file.IsCompressed {
		activityDetail = fmt.Sprintf("Compressed file downloaded (%.2f%% of original size)", file.CompressionRatio*100)
	}
	if file.IsSharded {
		activityDetail += fmt.Sprintf(" using Reed-Solomon reconstruction (%d data, %d parity shards)",
			file.DataShardCount, file.ParityShardCount)
	}

	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       user.ID,
		ActivityType: "download",
		FileID:       &file.ID,
		IPAddress:    ipAddress,
		Status:       "success",
		Details:      activityDetail,
	}); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}
}

func (c *DownloadFileController) sendFileResponse(ctx *gin.Context, file *models.File, data []byte) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	ctx.Header("Content-Type", file.MimeType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	log.Printf("Sending file response: %s (sharded=%v, compressed=%v)",
		file.Name, file.IsSharded, file.IsCompressed)
	ctx.Data(http.StatusOK, file.MimeType, data)
}
