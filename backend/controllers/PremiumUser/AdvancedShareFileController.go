package PremiumUser

import (
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
}

func NewShareFileController(
    fileModel *models.FileModel,
    fileShareModel *models.FileShareModel,
    keyFragmentModel *models.KeyFragmentModel,
    encryptionService *services.EncryptionService,
    activityLogModel *models.ActivityLogModel,
    rsService *services.ReedSolomonService,
) *ShareFileController {
    return &ShareFileController{
        fileModel:         fileModel,
        fileShareModel:    fileShareModel,
        keyFragmentModel:  keyFragmentModel,
        encryptionService: encryptionService,
        activityLogModel:  activityLogModel,
        rsService:         rsService,
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

    user, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "status": "error",
            "error":  "Unauthorized access",
        })
        return
    }
    currentUser := user.(*models.User)

    // Get and validate file
    file, err := c.fileModel.GetFileForDownload(req.FileID, currentUser.ID)
    if err != nil {
        log.Printf("File access error: %v", err)
        ctx.JSON(http.StatusNotFound, gin.H{
            "status": "error",
            "error":  "File not found or access denied",
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

    // Get key fragments
    fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
    if err != nil || len(fragments) == 0 {
        log.Printf("Failed to retrieve key fragments: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve key fragments",
        })
        return
    }

    // Encrypt fragment for sharing
    encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
        fragments[0].Data,
        []byte(req.Password),
    )
    if err != nil {
        log.Printf("Fragment encryption error: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to process share encryption",
        })
        return
    }

    // Create share record
    share := &models.FileShare{
        FileID:               file.ID,
        SharedBy:             currentUser.ID,
        EncryptedKeyFragment: encryptedFragment,
        ExpiresAt:            req.ExpiresAt,
        MaxDownloads:         req.MaxDownloads,
        IsActive:             true,
    }

    if err := c.fileShareModel.CreateFileShareWithStatus(share, req.Password); err != nil {
        log.Printf("Share creation error: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to create share",
        })
        return
    }

    // Log activity
    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       currentUser.ID,
        ActivityType: "share",
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

    // Validate share access with premium features
    share, err := c.fileShareModel.ValidateShareAccess(shareLink, req.Password)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "status": "error",
            "error":  err.Error(),
        })
        return
    }

    // Get file details first
    file, err := c.fileModel.GetFileByID(share.FileID)
    if err != nil {
        log.Printf("File retrieval error: %v", err)
        ctx.JSON(http.StatusNotFound, gin.H{
            "status": "error",
            "error":  "File not found",
        })
        return
    }

    // Get all key fragments
    allFragments, err := c.keyFragmentModel.GetKeyFragments(share.FileID)
    if err != nil || len(allFragments) == 0 {
        log.Printf("Failed to retrieve file fragments: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to process file access",
        })
        return
    }

    // Decrypt shared fragment
    decryptedFragment, err := c.encryptionService.DecryptKeyFragment(
        share.EncryptedKeyFragment,
        []byte(req.Password),
    )
    if err != nil {
        log.Printf("Fragment decryption error: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to process file decryption",
        })
        return
    }

    // Prepare key shares
    shares := make([]services.KeyShare, len(allFragments))
    for i, fragment := range allFragments {
        if i == 0 {
            shares[i] = services.KeyShare{
                Index: fragment.FragmentIndex,
                Value: string(decryptedFragment),
            }
        } else {
            shares[i] = services.KeyShare{
                Index: fragment.FragmentIndex,
                Value: string(fragment.Data),
            }
        }
    }

    // Get file data based on storage type
    var encryptedData []byte
    if file.IsSharded {
        log.Printf("Retrieving sharded data for file %d", file.ID)
        encryptedData, err = c.getShardedData(file)
    } else {
        log.Printf("Reading file content from path: %s", file.FilePath)
        encryptedData, err = c.fileModel.ReadFileContent(file.FilePath)
    }

    if err != nil {
        log.Printf("File data retrieval error: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to read file data",
        })
        return
    }

    // Decrypt file data
    decryptedData, err := c.encryptionService.DecryptFile(
        encryptedData,
        file.EncryptionIV,
        shares,
        int(file.Threshold),
        file.EncryptionSalt,
    )
    if err != nil {
        log.Printf("File decryption error: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to decrypt file",
        })
        return
    }

    // Update share status
    if err := c.fileShareModel.IncrementDownloadCount(share.ID); err != nil {
        log.Printf("Failed to increment download count: %v", err)
    }

    // Log activity
    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        ActivityType: "share_download",
        FileID:       &file.ID,
        IPAddress:    ctx.ClientIP(),
        Status:       "success",
        Details:      fmt.Sprintf("Premium share download (Sharded: %v)", file.IsSharded),
    }); err != nil {
        log.Printf("Failed to log share download activity: %v", err)
    }

    c.sendFileResponse(ctx, file, decryptedData)
}

func (c *ShareFileController) getShardedData(file *models.File) ([]byte, error) {
    // Retrieve shards
    fileShards, err := c.rsService.RetrieveShards(file.ID, int(file.DataShardCount+file.ParityShardCount))
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve shards: %w", err)
    }

    // Log shard information
    validShards := 0
    for i, shard := range fileShards.Shards {
        if shard != nil {
            validShards++
            log.Printf("Shard %d: %d bytes", i, len(shard))
        } else {
            log.Printf("Shard %d: Missing", i)
        }
    }

    // Validate available shards
    if !c.rsService.ValidateShards(fileShards.Shards, int(file.DataShardCount)) {
        return nil, fmt.Errorf("insufficient shards for reconstruction: have %d, need %d",
            validShards, file.DataShardCount)
    }

    // Reconstruct file
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