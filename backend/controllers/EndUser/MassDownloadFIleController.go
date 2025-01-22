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
    "sync"

    "github.com/gin-gonic/gin"
)

type MassDownloadFileController struct {
    fileModel          *models.FileModel
    keyFragmentModel   *models.KeyFragmentModel
    encryptionService  *services.EncryptionService
    activityLogModel   *models.ActivityLogModel
    compressionService *services.CompressionService
    rsService          *services.ReedSolomonService
    serverKeyModel     *models.ServerMasterKeyModel
}

type DownloadResult struct {
    FileName    string `json:"file_name"`
    Status      string `json:"status"`
    Error       string `json:"error,omitempty"`
    FileID      uint   `json:"file_id"`
    Size        int64  `json:"size,omitempty"`
}

func NewMassDownloadFileController(
    fileModel *models.FileModel,
    keyFragmentModel *models.KeyFragmentModel,
    encryptionService *services.EncryptionService,
    activityLogModel *models.ActivityLogModel,
    compressionService *services.CompressionService,
    rsService *services.ReedSolomonService,
    serverKeyModel *models.ServerMasterKeyModel,
) *MassDownloadFileController {
    return &MassDownloadFileController{
        fileModel:          fileModel,
        keyFragmentModel:   keyFragmentModel,
        encryptionService:  encryptionService,
        activityLogModel:   activityLogModel,
        compressionService: compressionService,
        rsService:          rsService,
        serverKeyModel:     serverKeyModel,
    }
}

func (c *MassDownloadFileController) MassDownload(ctx *gin.Context) {
    log.Printf("Starting mass file download request")

    // Get current user
    currentUser, err := c.getCurrentUser(ctx)
    if err != nil {
        return
    }

    // Parse file IDs
    fileIDs, err := c.parseFileIDs(ctx)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status": "error",
            "error":  fmt.Sprintf("Invalid file IDs: %v", err),
        })
        return
    }

    if len(fileIDs) == 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status": "error",
            "error":  "No files specified for download",
        })
        return
    }

    // Process files concurrently
    var wg sync.WaitGroup
    results := make(chan DownloadResult, len(fileIDs))
    semaphore := make(chan struct{}, 5) // Limit concurrent downloads

    for _, fileID := range fileIDs {
        wg.Add(1)
        go func(id uint) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore

            result := c.processDownload(ctx, currentUser, id)
            results <- result
        }(fileID)
    }

    // Wait for all downloads to complete
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    downloadResults := make([]DownloadResult, 0)
    totalFiles := 0
    totalSize := int64(0)
    successCount := 0

    for result := range results {
        downloadResults = append(downloadResults, result)
        totalFiles++
        if result.Status == "success" {
            successCount++
            totalSize += result.Size
        }
    }

    // Log activity
    c.logMassDownloadActivity(currentUser, successCount, totalSize, ctx.ClientIP())

    // Return results
    ctx.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": gin.H{
            "total_files":     totalFiles,
            "success_count":   successCount,
            "total_size":      totalSize,
            "download_status": downloadResults,
        },
    })
}

func (c *MassDownloadFileController) GetFile(ctx *gin.Context) {
    log.Printf("Starting single file download from mass download request")

    // Get current user
    currentUser, err := c.getCurrentUser(ctx)
    if err != nil {
        return
    }

    // Get file ID
    fileID, err := c.getFileID(ctx)
    if err != nil {
        return
    }

    // Get and validate file with user access check
    file, err := c.getAndValidateFile(ctx, currentUser.ID, fileID)
    if err != nil {
        return
    }

    log.Printf("Processing file ID: %d, IsSharded: %v, IsCompressed: %v",
        file.ID, file.IsSharded, file.IsCompressed)

    // Get key shares
    shares, err := c.getKeyShares(ctx, file)
    if err != nil {
        log.Printf("Failed to get key shares: %v", err)
        return
    }

    // Get encrypted data
    encryptedData, err := c.getFileData(ctx, file)
    if err != nil {
        return
    }

    // Decrypt data
    decryptedData, err := c.decryptData(ctx, file, encryptedData, shares)
    if err != nil {
        return
    }

    // Decompress if needed
    finalData, err := c.handleDecompression(ctx, file, decryptedData)
    if err != nil {
        return
    }

    // Log the download
    c.logDownloadActivity(currentUser, file, ctx.ClientIP())

    // Send the file
    c.sendFileResponse(ctx, file, finalData)
}
func (c *MassDownloadFileController) logDownloadActivity(user *models.User, file *models.File, ipAddress string) error {
    activityDetail := "File downloaded successfully"
    if file.IsCompressed {
        activityDetail = fmt.Sprintf("Compressed file downloaded (%.2f%% of original size)", file.CompressionRatio*100)
    }
    if file.IsSharded {
        activityDetail += fmt.Sprintf(" using Reed-Solomon reconstruction (%d data, %d parity shards)",
            file.DataShardCount, file.ParityShardCount)
    }

    return c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       user.ID,
        ActivityType: "download",
        FileID:       &file.ID,
        IPAddress:    ipAddress,
        Status:       "success",
        Details:      activityDetail,
    })
}
func (c *MassDownloadFileController) logBatchDownloadActivity(user *models.User, files []*models.File, ipAddress string) error {
    var totalSize int64
    var compressedCount, shardedCount int
    
    for _, file := range files {
        totalSize += file.Size
        if file.IsCompressed {
            compressedCount++
        }
        if file.IsSharded {
            shardedCount++
        }
    }

    activityDetail := fmt.Sprintf("Batch download of %d files (total size: %.2f MB)", 
        len(files), float64(totalSize)/1024/1024)
    
    if compressedCount > 0 {
        activityDetail += fmt.Sprintf(", %d compressed files", compressedCount)
    }
    if shardedCount > 0 {
        activityDetail += fmt.Sprintf(", %d sharded files", shardedCount)
    }

    return c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       user.ID,
        ActivityType: "download",
        IPAddress:    ipAddress,
        Status:       "success",
        Details:      activityDetail,
    })
}


func (c *MassDownloadFileController) processDownload(ctx *gin.Context, user *models.User, fileID uint) DownloadResult {
    result := DownloadResult{
        FileID: fileID,
        Status: "failed",
    }

    file, err := c.getAndValidateFile(ctx, user.ID, fileID)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to get file: %v", err)
        return result
    }

    result.FileName = file.OriginalName
    result.Size = file.Size

    // Get key shares
    shares, err := c.getKeyShares(ctx, file)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to get key shares: %v", err)
        return result
    }

    // Get file data
    encryptedData, err := c.getFileData(ctx, file)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to get file data: %v", err)
        return result
    }

    // Decrypt data
    decryptedData, err := c.decryptData(ctx, file, encryptedData, shares)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to decrypt data: %v", err)
        return result
    }

    // Decompress if needed
    _, err = c.handleDecompression(ctx, file, decryptedData)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to decompress data: %v", err)
        return result
    }

    result.Status = "success"
    return result
}

func (c *MassDownloadFileController) parseFileIDs(ctx *gin.Context) ([]uint, error) {
    var request struct {
        FileIDs []uint `json:"file_ids"`
    }

    if err := ctx.ShouldBindJSON(&request); err != nil {
        return nil, err
    }

    return request.FileIDs, nil
}

func (c *MassDownloadFileController) logMassDownloadActivity(user *models.User, fileCount int, totalSize int64, ipAddress string) {
    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       user.ID,
        ActivityType: "mass_download",
        IPAddress:    ipAddress,
        Status:       "success",
        Details:      fmt.Sprintf("Processed %d files (total size: %.2f MB)", fileCount, float64(totalSize)/1024/1024),
    }); err != nil {
        log.Printf("Failed to log mass download activity: %v", err)
    }
}

func (c *MassDownloadFileController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
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

func (c *MassDownloadFileController) getFileID(ctx *gin.Context) (uint, error) {
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

func (c *MassDownloadFileController) getAndValidateFile(ctx *gin.Context, userID, fileID uint) (*models.File, error) {
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

func (c *MassDownloadFileController) validateFileMetadata(file *models.File) error {
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

    return nil
}

func (c *MassDownloadFileController) getKeyShares(ctx *gin.Context, file *models.File) ([]services.KeyShare, error) {
    log.Printf("Retrieving key fragments for file %d", file.ID)

    serverKey, err := c.serverKeyModel.GetActive()
    if err != nil {
        log.Printf("Failed to retrieve server key: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve server key",
        })
        return nil, err
    }

    serverKeyData, err := c.serverKeyModel.GetServerKey(serverKey.KeyID)
    if err != nil {
        log.Printf("Failed to get server key data: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to get server key data",
        })
        return nil, err
    }

    currentUser, err := c.getCurrentUser(ctx)
    if err != nil {
        return nil, err
    }

    fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
    if err != nil {
        log.Printf("Failed to retrieve key fragments: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve key fragments",
        })
        return nil, err
    }

    shares := make([]services.KeyShare, len(fragments))
    for i, fragment := range fragments {
        var decryptedFragment []byte

        if fragment.KeyFragment.HolderType == models.ServerHolder {
            decryptedFragment, err = services.DecryptMasterKey(
                fragment.Data,
                serverKeyData,
                fragment.KeyFragment.EncryptionNonce[:12],
            )
        } else {
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

        normalizedFragment := make([]byte, 32)
        copy(normalizedFragment, decryptedFragment)

        shares[i] = services.KeyShare{
            Index:        fragment.KeyFragment.FragmentIndex,
            Value:        hex.EncodeToString(normalizedFragment),
            NodeIndex:    fragment.KeyFragment.NodeIndex,
            FragmentPath: fragment.KeyFragment.FragmentPath,
        }
    }

    return shares, nil
}

func (c *MassDownloadFileController) getFileData(ctx *gin.Context, file *models.File) ([]byte, error) {
    if file.IsSharded {
        return c.getShardedData(ctx, file)
    }

    data, err := os.ReadFile(file.FilePath)
    if err != nil {
        log.Printf("Failed to read file from disk: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to read file data",
        })
        return nil, err
    }

    return data, nil
}

func (c *MassDownloadFileController) getShardedData(ctx *gin.Context, file *models.File) ([]byte, error) {
    fileShards, err := c.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
    if err != nil {
        log.Printf("Failed to retrieve shards: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve file shards",
        })
        return nil, err
    }

    if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Insufficient shards for reconstruction",
        })
        return nil, fmt.Errorf("insufficient shards")
    }

    reconstructed, err := c.rsService.ReconstructFile(fileShards.Shards, int(file.DataShardCount), int(file.ParityShardCount))
    if err != nil {
        log.Printf("Failed to reconstruct file: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to reconstruct file",
        })
        return nil, err
    }

    return reconstructed, nil
}

func (c *MassDownloadFileController) decryptData(ctx *gin.Context, file *models.File, data []byte, shares []services.KeyShare) ([]byte, error) {
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

    return decrypted, nil
}

func (c *MassDownloadFileController) handleDecompression(ctx *gin.Context, file *models.File, data []byte) ([]byte, error) {
    if !file.IsCompressed {
        return data, nil
    }

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

func (c *MassDownloadFileController) sendFileResponse(ctx *gin.Context, file *models.File, data []byte) {
    ctx.Header("Content-Description", "File Transfer")
    ctx.Header("Content-Transfer-Encoding", "binary")
    ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
    ctx.Header("Content-Type", file.MimeType)
    ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))

    log.Printf("Sending file response: %s (sharded=%v, compressed=%v)",
        file.Name, file.IsSharded, file.IsCompressed)
    ctx.Data(http.StatusOK, file.MimeType, data)
}