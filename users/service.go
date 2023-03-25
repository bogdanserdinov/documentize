package users

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lukasjarosch/go-docx"
	exelize "github.com/xuri/excelize/v2"
	"github.com/zeebo/errs"

	"documentize/pkg/fileutils"
)

var ErrAlreadyGenerated = errors.New("file already generated")

type Config struct {
	ExportDataPath  string
	DocTemplatePath string
}

type Service struct {
	config Config

	db DB
}

func New(config Config, db DB) *Service {
	return &Service{
		config: config,
		db:     db,
	}
}

func (service *Service) Create(ctx context.Context, name, email string) error {
	user := User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Status:    StatusUngenerated,
		CreatedAt: time.Now().UTC(),
	}

	return service.db.Create(ctx, user)
}

func (service *Service) Get(ctx context.Context, id uuid.UUID) (User, error) {
	return service.db.Get(ctx, id)
}

func (service *Service) List(ctx context.Context) ([]User, error) {
	return service.db.List(ctx)
}

func (service *Service) UpdateStatus(ctx context.Context, id uuid.UUID) error {
	return service.db.UpdateStatus(ctx, id)
}

// Export creates export data for users.
func (service *Service) Export(ctx context.Context) (string, string, io.ReadSeekCloser, error) {
	file := exelize.NewFile()
	users, err := service.db.List(ctx)
	if err != nil {
		return "", "", nil, err
	}

	fields := []string{"ID", "Email", "Name", "Status", "Created At"}

	// header section.
	err = file.SetSheetRow(fileutils.DefaultSheetName, "A1", &fields)
	if err != nil {
		return "", "", nil, err
	}

	for i := 0; i < len(users); i++ {
		// in Excel line starts from 1.
		// first line in header.
		row := i + 2

		err = errs.Combine(
			file.SetCellValue(fileutils.DefaultSheetName, fmt.Sprintf("A%d", row), users[i].ID),
			file.SetCellValue(fileutils.DefaultSheetName, fmt.Sprintf("B%d", row), users[i].Email),
			file.SetCellValue(fileutils.DefaultSheetName, fmt.Sprintf("C%d", row), users[i].Name),
			file.SetCellValue(fileutils.DefaultSheetName, fmt.Sprintf("D%d", row), users[i].Status),
			file.SetCellValue(fileutils.DefaultSheetName, fmt.Sprintf("E%d", row), users[i].CreatedAt))
		if err != nil {
			return "", "", nil, err
		}
	}

	fileName := fileutils.NameForExportFiles("users", fileutils.ExcelExtension)
	err = os.MkdirAll(service.config.ExportDataPath, os.ModePerm)
	if err != nil {
		return "", "", nil, err
	}

	err = file.SaveAs(filepath.Join(service.config.ExportDataPath, fileName))
	if err != nil {
		return "", "", nil, err
	}

	reader, err := os.Open(filepath.Join(service.config.ExportDataPath, fileName))

	return fileName, filepath.Join(service.config.ExportDataPath, fileName), reader, err
}

// DeleteGeneratedFile deletes exported file.
func (service *Service) DeleteGeneratedFile(fileName string) error {
	return os.Remove(filepath.Join(service.config.ExportDataPath, fileName))
}

func (service *Service) GenerateDoc(ctx context.Context, id uuid.UUID) (string, string, io.ReadSeekCloser, error) {
	user, err := service.db.Get(ctx, id)
	if err != nil {
		return "", "", nil, err
	}

	if user.Status != StatusUngenerated {
		return "", "", nil, ErrAlreadyGenerated
	}

	replaceMap := docx.PlaceholderMap{
		"name":  user.Name,
		"email": user.Email,
	}

	doc, err := docx.Open(service.config.DocTemplatePath)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not open template file %v", err)
	}

	err = doc.ReplaceAll(replaceMap)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not replace template fields %v", err)
	}

	filename := fmt.Sprintf("document_%s.docs", user.Name)
	err = doc.WriteToFile(filepath.Join(service.config.ExportDataPath, filename))
	if err != nil {
		return "", "", nil, fmt.Errorf("could not save generated file %v", err)
	}

	reader, err := os.Open(filepath.Join(service.config.ExportDataPath, filename))
	if err != nil {
		return "", "", nil, fmt.Errorf("could not read generated file %v", err)
	}

	err = service.db.UpdateStatus(ctx, id)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not update users generated status %v", err)
	}

	return filename, filepath.Join(service.config.ExportDataPath, filename), reader, err
}
