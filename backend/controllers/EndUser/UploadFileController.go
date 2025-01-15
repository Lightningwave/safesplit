package EndUser

import (
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "safesplit/models"
    "safesplit/services"
    "strconv"

    "github.com/gin-gonic/gin"
)

type UploadFileController struct {
    fileModel          *models.FileModel
    userModel          *models.UserModel
    activityLogModel   *models.ActivityLogModel
    encryptionService  *services.EncryptionService
    shamirService      *services.ShamirService
    keyFragmentModel   *models.KeyFragmentModel
    compressionService *services.CompressionService
    folderModel        *models.FolderModel
    rsService          *services.ReedSolomonService
}

func NewFileController(
    fileModel *models.FileModel,
    userModel *models.UserModel,
    activityLogModel *models.ActivityLogModel,
    encryptionService *services.EncryptionService,
    shamirService *services.ShamirService,
    keyFragmentModel *models.KeyFragmentModel,
    compressionService *services.CompressionService,
    folderModel *models.FolderModel,
    rsService *services.ReedSolomonService,
) *UploadFileController {
    return &UploadFileController{
        fileModel:          fileModel,
        userModel:          userModel,
        activityLogModel:   activityLogModel,
        encryptionService:  encryptionService,
        shamirService:      shamirService,
        keyFragmentModel:   keyFragmentModel,
        compressionService: compressionService,
        folderModel:        folderModel,
        rsService:          rsService,
    }
}

// Validation methods
func (c *UploadFileController) validateShamirParameters(n, k int) error {
    if n < k {
        return fmt.Errorf("number of shares (n) must be greater than or equal to threshold (k)")
    }
    if k < 2 {
        return fmt.Errorf("threshold (k) must be at least 2")
    }
    if n > 10 {
        return fmt.Errorf("number of shares (n) cannot exceed 10")
    }
    return nil
}

func (c *UploadFileController) validateRSParameters(dataShards, parityShards int) error {
    if dataShards < 1 {
        return fmt.Errorf("data shards must be at least 1")
    }
    if parityShards < 1 {
        return fmt.Errorf("parity shards must be at least 1")
    }
    if dataShards+parityShards > 20 {
        return fmt.Errorf("total number of shards cannot exceed 20")
    }
    return nil
}

func (c *UploadFileController) validateParameters(n, k, dataShards, parityShards int) error {
    if err := c.validateShamirParameters(n, k); err != nil {
        return err
    }
    
    if err := c.validateRSParameters(dataShards, parityShards); err != nil {
        return err
    }
    
    return nil
}

type processedFile struct {
    compressed []byte
    encrypted  []byte
    iv         []byte
    salt       []byte
    shares     []services.KeyShare
    shards     [][]byte
    fileHash   string
    ratio      float64
}

func (c *UploadFileController) Upload(ctx *gin.Context) {
    log.Printf("Starting file upload request")

    user, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Unauthorized access"})
        return
    }

    currentUser, ok := user.(*models.User)
    if !ok {
        ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Invalid user data"})
        return
    }

    // Parse Reed-Solomon parameters
    dataShards, err := strconv.Atoi(ctx.DefaultPostForm("data_shards", "4"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid data shards value"})
        return
    }

    parityShards, err := strconv.Atoi(ctx.DefaultPostForm("parity_shards", "2"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid parity shards value"})
        return
    }

    // Parse Shamir parameters
    nShares, err := strconv.Atoi(ctx.DefaultPostForm("shares", "5"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid number of shares"})
        return
    }

    threshold, err := strconv.Atoi(ctx.DefaultPostForm("threshold", "3"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid threshold"})
        return
    }

    if err := c.validateParameters(nShares, threshold, dataShards, parityShards); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
        return
    }

    fileHeader, err := ctx.FormFile("file")
    if err != nil {
        log.Printf("Error receiving file: %v", err)
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "No file was provided"})
        return
    }

    // Check storage quota
    if !currentUser.HasAvailableStorage(fileHeader.Size) {
        ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Insufficient storage space"})
        return
    }

    // Process file upload
    processedFile, err := c.processFileUpload(fileHeader, nShares, threshold, dataShards, parityShards)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
        return
    }

    // Handle folder assignment
    folderID := c.handleFolderAssignment(ctx, currentUser)
    if folderID == nil {
        return
    }

    // Create file record
    encryptedFileName := base64.RawURLEncoding.EncodeToString([]byte(fileHeader.Filename))
    fileRecord := &models.File{
        UserID:           currentUser.ID,
        FolderID:         folderID,
        Name:             encryptedFileName,
        OriginalName:     fileHeader.Filename,
        Size:             fileHeader.Size,
        CompressedSize:   int64(len(processedFile.compressed)),
        MimeType:         fileHeader.Header.Get("Content-Type"),
        EncryptionIV:     processedFile.iv,
        EncryptionSalt:   processedFile.salt,
        FileHash:         processedFile.fileHash,
        ShareCount:       uint(nShares),
        Threshold:        uint(threshold),
        DataShardCount:   uint(dataShards),
        ParityShardCount: uint(parityShards),
        IsCompressed:     true,
        IsSharded:        true,
        CompressionRatio: processedFile.ratio,
    }

    // Create file with shards
    if err := c.fileModel.CreateFileWithShards(fileRecord, processedFile.shares, processedFile.shards, c.keyFragmentModel); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "error":  fmt.Sprintf("Failed to save file information: %v", err),
        })
        return
    }

    // Log activity
    if err := c.activityLogModel.LogActivity(&models.ActivityLog{
        UserID:       currentUser.ID,
        ActivityType: "upload",
        FileID:       &fileRecord.ID,
        IPAddress:    ctx.ClientIP(),
        Status:       "success",
        Details:      fmt.Sprintf("File uploaded with %d data shards, %d parity shards, compressed to %.2f%%", 
            dataShards, parityShards, processedFile.ratio*100),
    }); err != nil {
        log.Printf("Failed to log activity: %v", err)
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "File uploaded successfully",
        "data": gin.H{
            "file": fileRecord,
            "shardInfo": gin.H{
                "dataShards":   dataShards,
                "parityShards": parityShards,
                "totalShards":  dataShards + parityShards,
            },
            "compressionStats": gin.H{
                "originalSize":     fileRecord.Size,
                "compressedSize":   fileRecord.CompressedSize,
                "compressionRatio": fmt.Sprintf("%.2f%%", processedFile.ratio*100),
            },
            "folder_id": folderID,
        },
    })
}

func (c *UploadFileController) processFileUpload(
    fileHeader *multipart.FileHeader, 
    n, k int,
    dataShards, parityShards int,
) (*processedFile, error) {
    log.Printf("Starting file processing - Size: %d bytes", fileHeader.Size)
    
    src, err := fileHeader.Open()
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    defer src.Close()

    content, err := io.ReadAll(src)
    if err != nil {
        return nil, fmt.Errorf("failed to read file content: %w", err)
    }
    log.Printf("Read file content - Size: %d bytes", len(content))

    // Calculate hash
    hash := sha256.Sum256(content)
    fileHash := base64.StdEncoding.EncodeToString(hash[:])
    log.Printf("File hash: %s", fileHash)

    // Compress content
    compressed, ratio, err := c.compressionService.Compress(content)
    if err != nil {
        return nil, fmt.Errorf("compression failed: %w", err)
    }
    log.Printf("Compressed data - Original: %d, Compressed: %d bytes, Ratio: %.2f%%",
        len(content), len(compressed), ratio*100)

    // Encrypt the compressed content
    encrypted, iv, salt, shares, err := c.encryptionService.EncryptFile(compressed, n, k)
    if err != nil {
        return nil, fmt.Errorf("encryption failed: %w", err)
    }
    log.Printf("Encrypted data - Size: %d bytes, Shares: %d", len(encrypted), len(shares))

    // Create Reed-Solomon shards
    fileShards, err := c.rsService.SplitFile(encrypted, dataShards, parityShards)
    if err != nil {
        return nil, fmt.Errorf("reed-solomon encoding failed: %w", err)
    }
    log.Printf("Created %d shards", len(fileShards.Shards))

    // Verify shard sizes
    shardSize := len(fileShards.Shards[0])
    for i, shard := range fileShards.Shards {
        if len(shard) != shardSize {
            return nil, fmt.Errorf("inconsistent shard size at index %d: got %d, want %d",
                i, len(shard), shardSize)
        }
        log.Printf("Shard %d size: %d bytes", i, len(shard))
    }

    return &processedFile{
        compressed: compressed,
        encrypted:  encrypted,
        iv:         iv,
        salt:       salt,
        shares:     shares,
        shards:     fileShards.Shards,
        fileHash:   fileHash,
        ratio:      ratio,
    }, nil
}

// Helper method to handle folder assignment
func (c *UploadFileController) handleFolderAssignment(ctx *gin.Context, currentUser *models.User) *uint {
    if folderIDStr := ctx.PostForm("folder_id"); folderIDStr != "" {
        id, err := strconv.ParseUint(folderIDStr, 10, 32)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid folder ID format"})
            return nil
        }
        parsedID := uint(id)
        return &parsedID
    }

    folders, err := c.folderModel.GetUserFolders(currentUser.ID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to check folders"})
        return nil
    }

    var myFilesID *uint
    for _, folder := range folders {
        if folder.Name == "My Files" {
            myFilesID = &folder.ID
            break
        }
    }

    if myFilesID == nil {
        defaultFolder := &models.Folder{
            UserID: currentUser.ID,
            Name:   "My Files",
        }
        if err := c.folderModel.CreateFolder(defaultFolder); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to create default folder"})
            return nil
        }
        myFilesID = &defaultFolder.ID
    }

    return myFilesID
}