package EndUser

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
    "strings"
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
    FileID   uint   `json:"file_id" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
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

    // Get current user
    user, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "status": "error",
            "error":  "Unauthorized access",
        })
        return
    }
    currentUser := user.(*models.User)

    // Verify file ownership
    file, err := c.fileModel.GetFileForDownload(req.FileID, currentUser.ID)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{
            "status": "error",
            "error":  "File not found or access denied",
        })
        return
    }

    // Get key fragments
    fragments, err := c.keyFragmentModel.GetKeyFragments(file.ID)
    if err != nil || len(fragments) == 0 {
        log.Printf("Failed to retrieve key fragments for file %d: %v", file.ID, err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to retrieve key fragments",
        })
        return
    }

    log.Printf("Creating share for file %d with %d fragments", file.ID, len(fragments))

    // Encrypt fragment with share password
    encryptedFragment, err := c.encryptionService.EncryptKeyFragment(
        fragments[0].EncryptedFragment,
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

    // Create share record
    share := &models.FileShare{
        FileID:               file.ID,
        SharedBy:             currentUser.ID,
        EncryptedKeyFragment: encryptedFragment,
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

    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       currentUser.ID,
        ActivityType: "share",
        FileID:       &file.ID,
        IPAddress:    ctx.ClientIP(),
        Status:       "success",
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

    // Validate share access
    share, err := c.fileShareModel.ValidateShare(shareLink, req.Password)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "status": "error",
            "error":  "Invalid share link or password",
        })
        return
    }

    log.Printf("Processing share access for share link: %s", shareLink)

    // Get file details first to validate metadata
    file, err := c.fileModel.GetFileByID(share.FileID)
    if err != nil {
        log.Printf("Failed to get file %d: %v", share.FileID, err)
        ctx.JSON(http.StatusNotFound, gin.H{
            "status": "error",
            "error":  "File not found",
        })
        return
    }

    // Get file fragments
    fragments, err := c.keyFragmentModel.GetKeyFragments(share.FileID)
    if err != nil || len(fragments) == 0 {
        log.Printf("Failed to get key fragments for file %d: %v", share.FileID, err)
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
        log.Printf("Failed to decrypt key fragment: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to process file decryption",
        })
        return
    }

    // Prepare key shares
    shares := make([]services.KeyShare, len(fragments))
    for i, fragment := range fragments {
        if i == 0 {
            shares[i] = services.KeyShare{
                Index: fragment.FragmentIndex,
                Value: string(decryptedFragment),
            }
        } else {
            shares[i] = services.KeyShare{
                Index: fragment.FragmentIndex,
                Value: string(fragment.EncryptedFragment),
            }
        }
    }

    // Get encrypted data based on storage type
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

    // Decrypt the data
    decryptedData, err := c.encryptionService.DecryptFile(
        encryptedData,
        file.EncryptionIV,
        shares,
        int(file.Threshold),
        file.EncryptionSalt,
    )
    if err != nil {
        log.Printf("Failed to decrypt file data: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  "Failed to decrypt file",
        })
        return
    }

    // Log share access
    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       share.SharedBy,
        ActivityType: "download",
        FileID:       &file.ID,
        IPAddress:    ctx.ClientIP(),
        Status:       "success",
        Details:      fmt.Sprintf("Shared file download (Sharded: %v)", file.IsSharded),
    }); err != nil {
        log.Printf("Failed to log share download activity: %v", err)
    }

    // Send file response
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
    ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
        sanitizedFilename,
        encodedFilename))
    ctx.Header("Content-Type", file.MimeType)
    ctx.Header("Content-Length", fmt.Sprintf("%d", len(data)))

    log.Printf("Sending file response: %s (Size: %d bytes)", file.OriginalName, len(data))
    ctx.Data(http.StatusOK, file.MimeType, data)
}