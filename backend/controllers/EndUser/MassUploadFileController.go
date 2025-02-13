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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type MassUploadFileController struct {
	fileModel          *models.FileModel
	userModel          *models.UserModel
	activityLogModel   *models.ActivityLogModel
	encryptionService  *services.EncryptionService
	shamirService      *services.ShamirService
	keyFragmentModel   *models.KeyFragmentModel
	compressionService *services.CompressionService
	folderModel        *models.FolderModel
	rsService          *services.ReedSolomonService
	serverKeyModel     *models.ServerMasterKeyModel
}

type massProcessedFile struct {
	compressed []byte
	encrypted  []byte
	iv         []byte
	salt       []byte
	shares     []services.KeyShare
	shards     [][]byte
	fileHash   string
	ratio      float64
}
type UploadParams struct {
	EncryptionType services.EncryptionType
	NShares        int
	Threshold      int
	DataShards     int
	ParityShards   int
}

type UploadResult struct {
	FileName    string      `json:"file_name"`
	Status      string      `json:"status"`
	Error       string      `json:"error,omitempty"`
	FileID      uint        `json:"file_id,omitempty"`
	Size        int64       `json:"size,omitempty"`
	FileInfo    interface{} `json:"file_info,omitempty"`
	StoragePath string      `json:"storage_path,omitempty"`
}

func NewMassUploadFileController(
	fileModel *models.FileModel,
	userModel *models.UserModel,
	activityLogModel *models.ActivityLogModel,
	encryptionService *services.EncryptionService,
	shamirService *services.ShamirService,
	keyFragmentModel *models.KeyFragmentModel,
	compressionService *services.CompressionService,
	folderModel *models.FolderModel,
	rsService *services.ReedSolomonService,
	serverKeyModel *models.ServerMasterKeyModel,
) *MassUploadFileController {
	return &MassUploadFileController{
		fileModel:          fileModel,
		userModel:          userModel,
		activityLogModel:   activityLogModel,
		encryptionService:  encryptionService,
		shamirService:      shamirService,
		keyFragmentModel:   keyFragmentModel,
		compressionService: compressionService,
		folderModel:        folderModel,
		rsService:          rsService,
		serverKeyModel:     serverKeyModel,
	}
}

func (c *MassUploadFileController) MassUpload(ctx *gin.Context) {
	log.Printf("Starting mass file upload request")

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

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Failed to parse form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "No files provided"})
		return
	}

	encryptionType := services.EncryptionType(ctx.DefaultPostForm("encryption_type", string(services.StandardEncryption)))
	if err := c.validateEncryptionType(encryptionType, currentUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	nShares := c.parseIntParam(ctx, "shares", 5)
	threshold := c.parseIntParam(ctx, "threshold", 3)
	dataShards := c.parseIntParam(ctx, "data_shards", 4)
	parityShards := c.parseIntParam(ctx, "parity_shards", 2)

	if err := c.validateParameters(nShares, threshold, dataShards, parityShards); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Handle folder assignment
	folderID := c.handleFolderAssignment(ctx, currentUser)
	if folderID == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid folder assignment"})
		return
	}

	// Calculate total size
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	// Check storage quota
	if !currentUser.HasAvailableStorage(totalSize) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Insufficient storage space"})
		return
	}

	// Process files concurrently
	var wg sync.WaitGroup
	results := make(chan UploadResult, len(files))
	semaphore := make(chan struct{}, 5) // Limit concurrent uploads

	uploadParams := &UploadParams{
		EncryptionType: encryptionType,
		NShares:        nShares,
		Threshold:      threshold,
		DataShards:     dataShards,
		ParityShards:   parityShards,
	}

	for _, fileHeader := range files {
		wg.Add(1)
		go func(fh *multipart.FileHeader) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := c.processUpload(ctx, fh, currentUser, folderID, uploadParams)
			results <- result
		}(fileHeader)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	uploadResults := make([]UploadResult, 0, len(files))
	for result := range results {
		uploadResults = append(uploadResults, result)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Mass file upload processing complete",
		"results": uploadResults,
	})
}

func (c *MassUploadFileController) processUpload(
	ctx *gin.Context,
	fileHeader *multipart.FileHeader,
	user *models.User,
	folderID *uint,
	params *UploadParams,
) UploadResult {
	result := UploadResult{
		FileName: fileHeader.Filename,
		Status:   "failed",
	}

	// Process file upload
	processedFile, err := c.processFileUpload(
		fileHeader,
		params.NShares,
		params.Threshold,
		params.DataShards,
		params.ParityShards,
		params.EncryptionType,
	)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Get server key
	serverKey, err := c.serverKeyModel.GetActive()
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get server key: %v", err)
		return result
	}

	// Create file record
	fileRecord, err := c.createFileRecord(fileHeader, user.ID, folderID, processedFile, params, serverKey)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create file record: %v", err)
		return result
	}

	// Create file with shards
	if err := c.fileModel.CreateFileWithShards(
		fileRecord,
		processedFile.shares,
		processedFile.shards,
		c.keyFragmentModel,
		c.serverKeyModel,
	); err != nil {
		result.Error = fmt.Sprintf("Failed to save file: %v", err)
		return result
	}

	// Log activity
	c.logUploadActivity(user.ID, fileRecord, ctx.ClientIP(), params)

	// Update result
	result.Status = "success"
	result.FileID = fileRecord.ID
	result.Size = fileRecord.Size
	result.FileInfo = c.getFileInfo(fileRecord, params)

	return result
}

func (c *MassUploadFileController) processFileUpload(
	fileHeader *multipart.FileHeader,
	n, k int,
	dataShards, parityShards int,
	encType services.EncryptionType,
) (*processedFile, error) {
	log.Printf("Starting file processing - Size: %d bytes, Encryption: %s", fileHeader.Size, encType)

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

	// Get active server key
	serverKey, err := c.serverKeyModel.GetActive()
	if err != nil {
		return nil, fmt.Errorf("failed to get server key: %w", err)
	}

	// Generate a temporary file ID for encryption
	tempFileID := uint(time.Now().UnixNano())

	// Use EncryptFileWithType for encryption
	encrypted, iv, salt, shares, err := c.encryptionService.EncryptFileWithType(
		compressed,
		n,
		k,
		tempFileID,
		serverKey.KeyID,
		encType,
	)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Create Reed-Solomon shards
	fileShards, err := c.rsService.SplitFile(encrypted, dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon encoding failed: %w", err)
	}
	log.Printf("Created %d shards", len(fileShards.Shards))

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

func (c *MassUploadFileController) createFileRecord(
	fileHeader *multipart.FileHeader,
	userID uint,
	folderID *uint,
	processedFile *processedFile,
	params *UploadParams,
	serverKey *models.ServerMasterKey,
) (*models.File, error) {
	if processedFile == nil {
		return nil, fmt.Errorf("processedFile is nil")
	}

	if serverKey == nil {
		return nil, fmt.Errorf("serverKey is nil")
	}

	encryptedFileName := base64.RawURLEncoding.EncodeToString([]byte(fileHeader.Filename))

	return &models.File{
		UserID:            userID,
		FolderID:          folderID,
		Name:              encryptedFileName,
		OriginalName:      fileHeader.Filename,
		Size:              fileHeader.Size,
		CompressedSize:    int64(len(processedFile.compressed)),
		MimeType:          fileHeader.Header.Get("Content-Type"),
		EncryptionIV:      processedFile.iv,
		EncryptionSalt:    processedFile.salt,
		EncryptionType:    params.EncryptionType,
		EncryptionVersion: 1,
		FileHash:          processedFile.fileHash,
		ShareCount:        uint(params.NShares),
		Threshold:         uint(params.Threshold),
		DataShardCount:    uint(params.DataShards),
		ParityShardCount:  uint(params.ParityShards),
		IsCompressed:      true,
		IsSharded:         true,
		CompressionRatio:  processedFile.ratio,
		ServerKeyID:       serverKey.KeyID,
		MasterKeyVersion:  1,
	}, nil
}

func (c *MassUploadFileController) validateEncryptionType(encType services.EncryptionType, user *models.User) error {
	switch encType {
	case services.StandardEncryption:
		return nil
	case services.ChaCha20, services.Twofish:
		if !user.IsPremiumUser() {
			return fmt.Errorf("%s encryption requires a premium account", encType)
		}
		return nil
	default:
		return fmt.Errorf("unsupported encryption type: %s", encType)
	}
}

func (c *MassUploadFileController) validateParameters(n, k, dataShards, parityShards int) error {
	if n < k {
		return fmt.Errorf("number of shares (n) must be greater than or equal to threshold (k)")
	}
	if k < 2 {
		return fmt.Errorf("threshold (k) must be at least 2")
	}
	if n > 10 {
		return fmt.Errorf("number of shares (n) cannot exceed 10")
	}
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

func (c *MassUploadFileController) parseIntParam(ctx *gin.Context, name string, defaultValue int) int {
	value, err := strconv.Atoi(ctx.DefaultPostForm(name, fmt.Sprintf("%d", defaultValue)))
	if err != nil {
		return defaultValue
	}
	return value
}

func (c *MassUploadFileController) handleFolderAssignment(ctx *gin.Context, user *models.User) *uint {
	if folderIDStr := ctx.PostForm("folder_id"); folderIDStr != "" {
		id, err := strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			log.Printf("Invalid folder ID format: %v", err)
			return nil
		}
		parsedID := uint(id)

		folder, err := c.folderModel.GetFolderByID(parsedID, user.ID)
		if err != nil {
			log.Printf("Folder not found or access denied: %v", err)
			return nil
		}
		return &folder.ID
	}

	folders, err := c.folderModel.GetUserFolders(user.ID)
	if err != nil {
		log.Printf("Failed to get user folders: %v", err)
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
			UserID: user.ID,
			Name:   "My Files",
		}
		if err := c.folderModel.CreateFolder(defaultFolder); err != nil {
			log.Printf("Failed to create default folder: %v", err)
			return nil
		}
		myFilesID = &defaultFolder.ID
	}

	return myFilesID
}

func (c *MassUploadFileController) logUploadActivity(userID uint, file *models.File, ipAddress string, params *UploadParams) {
	if err := c.activityLogModel.LogActivity(&models.ActivityLog{
		UserID:       userID,
		ActivityType: "upload",
		FileID:       &file.ID,
		IPAddress:    ipAddress,
		Status:       "success",
		Details: fmt.Sprintf(
			"File uploaded with %s encryption, %d shares, %d threshold, %.2f%% compression",
			params.EncryptionType,
			params.NShares,
			params.Threshold,
			file.CompressionRatio*100,
		),
	}); err != nil {
		log.Printf("Failed to log upload activity: %v", err)
	}
}

func (c *MassUploadFileController) getFileInfo(file *models.File, params *UploadParams) gin.H {
	return gin.H{
		"id":   file.ID,
		"name": file.Name,
		"size": file.Size,
		"compression": gin.H{
			"original_size":   file.Size,
			"compressed_size": file.CompressedSize,
			"ratio":           fmt.Sprintf("%.2f%%", file.CompressionRatio*100),
		},
		"encryption": gin.H{
			"type":    params.EncryptionType,
			"version": file.EncryptionVersion,
		},
		"sharding": gin.H{
			"data_shards":   params.DataShards,
			"parity_shards": params.ParityShards,
			"total_shards":  params.DataShards + params.ParityShards,
		},
		"creation_time": file.CreatedAt,
		"file_hash":     file.FileHash,
	}
}

type UploadError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *UploadError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func newUploadError(code, message string, details string) *UploadError {
	return &UploadError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func (c *MassUploadFileController) calculateBatchSize(fileCount int) int {

	if fileCount <= 5 {
		return fileCount
	}
	if fileCount <= 20 {
		return 5
	}
	return 10
}

func (c *MassUploadFileController) validateTotalSize(files []*multipart.FileHeader) (int64, error) {
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
		if file.Size > 2<<30 {
			return 0, newUploadError(
				"FILE_TOO_LARGE",
				fmt.Sprintf("File %s exceeds maximum size limit", file.Filename),
				"Maximum file size is 2GB",
			)
		}
	}

	// Total batch size limit (e.g., 10GB)
	if totalSize > 10<<30 {
		return 0, newUploadError(
			"BATCH_TOO_LARGE",
			"Total upload size exceeds maximum batch limit",
			"Maximum batch size is 10GB",
		)
	}

	return totalSize, nil
}

func (c *MassUploadFileController) GetUploadStats() gin.H {
	return gin.H{
		"max_file_size":       "2GB",
		"max_batch_size":      "10GB",
		"max_files_per_batch": 100,
		"supported_types": []string{
			"application/pdf",
			"image/*",
			"text/*",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.*",
		},
	}
}