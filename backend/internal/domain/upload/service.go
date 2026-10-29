package upload

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidFileType     = errors.New("unsupported or disallowed file format")
	ErrFileTooLarge        = errors.New("file size exceeds maximum allowed limit")
	ErrEmptyFile           = errors.New("uploaded file is empty")
	ErrUnauthorizedAccess  = errors.New("unauthorized access to file")
)

const (
	MaxPhotoSize    = 15 * 1024 * 1024 // 15 MB
	MaxDocumentSize = 30 * 1024 * 1024 // 30 MB
	MaxExtractSize  = 2 * 1024 * 1024  // 2 MB max extract for context
)

var allowedPhotoExts = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".gif":  "image/gif",
}

var allowedDocExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".md":   true,
	".csv":  true,
	".json": true,
	".py":   true,
	".js":   true,
	".ts":   true,
	".tsx":  true,
	".jsx":  true,
	".go":   true,
	".rs":   true,
	".c":    true,
	".cpp":  true,
	".h":    true,
	".hpp":  true,
	".html": true,
	".css":  true,
	".yaml": true,
	".yml":  true,
	".sh":   true,
	".sql":  true,
	".xml":  true,
	".java": true,
	".kt":   true,
	".swift": true,
	".rb":   true,
	".php":  true,
}

type FileUploadResult struct {
	File     *UploadedFile `json:"file"`
	URL      string        `json:"url"`
	Content  string        `json:"content,omitempty"`
}

type UploadService struct {
	repo       UploadRepository
	storageDir string
}

func NewUploadService(repo UploadRepository, storageDir ...string) *UploadService {
	dir := "uploads"
	if len(storageDir) > 0 && storageDir[0] != "" {
		dir = storageDir[0]
	}

	_ = os.MkdirAll(dir, 0755)

	return &UploadService{
		repo:       repo,
		storageDir: dir,
	}
}

func (s *UploadService) SaveUploadedFile(
	ctx context.Context,
	userID uuid.UUID,
	fileHeader *multipart.FileHeader,
	requestedType string, // "photo" or "file" or auto
) (*FileUploadResult, error) {
	if fileHeader == nil || fileHeader.Size == 0 {
		return nil, ErrEmptyFile
	}

	filename := filepath.Base(fileHeader.Filename)
	ext := strings.ToLower(filepath.Ext(filename))

	// Detect and validate type
	var isPhoto bool
	var mimeType string

	photoMime, isPhotoExt := allowedPhotoExts[ext]
	isDocExt := allowedDocExts[ext]

	if requestedType == "photo" {
		if !isPhotoExt {
			return nil, fmt.Errorf("%w: photos must strictly be PNG, JPG, JPEG, WEBP, or GIF", ErrInvalidFileType)
		}
		if fileHeader.Size > MaxPhotoSize {
			return nil, fmt.Errorf("%w: max image size is 15MB", ErrFileTooLarge)
		}
		isPhoto = true
		mimeType = photoMime
	} else if requestedType == "file" {
		if !isDocExt {
			return nil, fmt.Errorf("%w: unsupported file extension '%s'", ErrInvalidFileType, ext)
		}
		if fileHeader.Size > MaxDocumentSize {
			return nil, fmt.Errorf("%w: max document size is 30MB", ErrFileTooLarge)
		}
		isPhoto = false
		mimeType = detectMimeType(ext, fileHeader.Header.Get("Content-Type"))
	} else {
		// Auto-detect based on extension
		if isPhotoExt {
			isPhoto = true
			mimeType = photoMime
			if fileHeader.Size > MaxPhotoSize {
				return nil, fmt.Errorf("%w: max image size is 15MB", ErrFileTooLarge)
			}
		} else if isDocExt {
			isPhoto = false
			mimeType = detectMimeType(ext, fileHeader.Header.Get("Content-Type"))
			if fileHeader.Size > MaxDocumentSize {
				return nil, fmt.Errorf("%w: max document size is 30MB", ErrFileTooLarge)
			}
		} else {
			return nil, fmt.Errorf("%w: extension '%s' is not supported", ErrInvalidFileType, ext)
		}
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer srcFile.Close()

	// Read content for storage and context extraction
	data, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Persist to disk
	fileID := uuid.New()
	safeFilename := fmt.Sprintf("%s%s", fileID.String(), ext)
	diskPath := filepath.Join(s.storageDir, safeFilename)

	if err := os.WriteFile(diskPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to persist file to disk: %w", err)
	}

	fileRecord := &UploadedFile{
		ID:        fileID,
		UserID:    userID,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      fileHeader.Size,
		FilePath:  diskPath,
		FileType:  "document",
		CreatedAt: timeNowUTC(),
	}

	if isPhoto {
		fileRecord.FileType = "image"
	}

	if err := s.repo.Create(ctx, fileRecord); err != nil {
		_ = os.Remove(diskPath)
		return nil, fmt.Errorf("failed to save file record in database: %w", err)
	}

	var extractedContent string
	if isPhoto {
		// Base64 encode for vision ingestion
		extractedContent = base64.StdEncoding.EncodeToString(data)
	} else {
		// Extract text content for LLM context
		extractedContent = extractTextContent(data, ext)
	}

	return &FileUploadResult{
		File:    fileRecord,
		URL:     fmt.Sprintf("/api/v1/upload/%s/file", fileID.String()),
		Content: extractedContent,
	}, nil
}

func (s *UploadService) GetFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*UploadedFile, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}
	return file, nil
}

func (s *UploadService) GetFileContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (string, error) {
	file, err := s.GetFile(ctx, id, userID)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		return "", err
	}

	if file.FileType == "image" {
		return base64.StdEncoding.EncodeToString(data), nil
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	return extractTextContent(data, ext), nil
}

func detectMimeType(ext, headerMime string) string {
	if headerMime != "" && headerMime != "application/octet-stream" {
		return headerMime
	}
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js", ".ts", ".tsx", ".jsx":
		return "text/javascript"
	case ".py":
		return "text/x-python"
	case ".go":
		return "text/x-go"
	case ".rs":
		return "text/x-rust"
	default:
		return "text/plain"
	}
}

func extractTextContent(data []byte, ext string) string {
	if len(data) > MaxExtractSize {
		data = data[:MaxExtractSize]
	}

	switch ext {
	case ".pdf":
		// Basic printable ASCII/UTF-8 stream extraction for PDF text chunks
		var b strings.Builder
		inText := false
		for i := 0; i < len(data); i++ {
			c := data[i]
			if c == '(' && !inText {
				inText = true
				continue
			}
			if c == ')' && inText {
				inText = false
				b.WriteByte(' ')
				continue
			}
			if inText && (c >= 32 && c <= 126 || c == '\n' || c == '\t') {
				b.WriteByte(c)
			}
		}
		res := strings.TrimSpace(b.String())
		if len(res) > 50 {
			return res
		}
		return string(bytes.ToValidUTF8(data, []byte(" ")))
	default:
		// Text/code/markdown/json files
		return string(bytes.ToValidUTF8(data, []byte("")))
	}
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}
