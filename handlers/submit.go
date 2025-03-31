package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/models"
)

func UploadTwoFilesHandler(c *fiber.Ctx) error {
	// Lấy file .cpp
	cppFile, err := c.FormFile("cpp_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể đọc file .cpp từ request",
		})
	}
	if !strings.HasSuffix(cppFile.Filename, ".cpp") {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "File cpp_file phải có đuôi .cpp",
		})
	}

	// Lấy file .h
	hFile, err := c.FormFile("h_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể đọc file .h từ request",
		})
	}
	if !strings.HasSuffix(hFile.Filename, ".h") {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "File h_file phải có đuôi .h",
		})
	}

	// Tạo base ID ngẫu nhiên 7 chữ số
	baseID := generateFileID()
	cppFileID := fmt.Sprintf("[%s]cpp", baseID) // [xxxxxxx]cpp
	hFileID := fmt.Sprintf("[%s]h", baseID)     // [xxxxxxx]h

	// Gọi UploadFileToJobeHandler cho file .cpp
	cppCtx := c.Context() // Tạo context mới để gọi hàm
	cppCtx.Set("form_file_key", "cpp_file")
	cppCtx.Set("file_id", cppFileID)
	if err := UploadFileToJobeHandler(cppCtx); err != nil {
		return err // Trả về lỗi nếu upload .cpp thất bại
	}

	// Gọi UploadFileToJobeHandler cho file .h
	hCtx := c.Context() // Tạo context mới để gọi hàm
	hCtx.Set("form_file_key", "h_file")
	hCtx.Set("file_id", hFileID)
	if err := UploadFileToJobeHandler(hCtx); err != nil {
		return err // Trả về lỗi nếu upload .h thất bại
	}

	// Trả về kết quả thành công cho cả hai file
	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Cả hai file đã được upload thành công",
		"cpp_file_id": cppFileID,
		"h_file_id":   hFileID,
	})
}
