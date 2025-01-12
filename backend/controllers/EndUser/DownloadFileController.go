package EndUser

import (
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
}

func NewDownloadFileController(
    fileModel *models.FileModel,
    keyFragmentModel *models.KeyFragmentModel,
    encryptionService *services.EncryptionService,
    activityLogModel *models.ActivityLogModel,
    compressionService *services.CompressionService,
    rsService *services.ReedSolomonService,
) *DownloadFileController {
    return &DownloadFileController{
        fileModel:          fileModel,
        keyFragmentModel:   keyFragmentModel,
        encryptionService:  encryptionService,
        activityLogModel:   activityLogModel,
        compressionService: compressionService,
        rsService:          rsService,
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
    if len(file.EncryptionIV) != 16 {
        return fmt.Errorf("invalid IV length: got %d, expected 16", len(file.EncryptionIV))
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

func (c *DownloadFileController) getFileData(ctx *gin.Context, file *models.File) ([]byte, error) {
    if file.IsSharded {
        log.Printf("Retrieving sharded file data for ID: %d", file.ID)
        return c.getShardedData(ctx, file)
    }

    log.Printf("Reading file from path: %s", file.FilePath)
    data, err := os.ReadFile(file.FilePath)
    if err != nil {
        log.Printf("Failed to read file: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to read file",
        })
        return nil, err
    }
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

func (c *DownloadFileController) getKeyShares(ctx *gin.Context, file *models.File) ([]services.KeyShare, error) {
    fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
    if err != nil {
        log.Printf("Failed to retrieve key fragments: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve key fragments",
        })
        return nil, err
    }

    if uint(len(fragments)) < file.ShareCount {
        log.Printf("Insufficient key fragments: found %d, expected %d", len(fragments), file.ShareCount)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "File key fragments are incomplete",
        })
        return nil, fmt.Errorf("insufficient key fragments")
    }

    shares := make([]services.KeyShare, len(fragments))
    for i, fragment := range fragments {
        shares[i] = services.KeyShare{
            Index: fragment.FragmentIndex,
            Value: string(fragment.EncryptedFragment),
        }
    }

    return shares, nil
}

func (c *DownloadFileController) decryptData(ctx *gin.Context, file *models.File, data []byte, shares []services.KeyShare) ([]byte, error) {
    log.Printf("Starting decryption for file ID: %d", file.ID)
    log.Printf("Data size: %d bytes", len(data))
    log.Printf("IV length: %d bytes", len(file.EncryptionIV))
    log.Printf("Salt length: %d bytes", len(file.EncryptionSalt))

    for i, share := range shares {
        log.Printf("Share %d - Index: %d, Value length: %d", i, share.Index, len(share.Value))
    }

    decrypted, err := c.encryptionService.DecryptFile(data, file.EncryptionIV, shares, int(file.Threshold), file.EncryptionSalt)
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