package common

import (
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/ledongthuc/pdf"
	"github.com/rs/zerolog/log"
)

func ExtractPDFText(file *os.File) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Error().Err(err).Msg("Failed to seek PDF file to start")
		return "", err
	}

	stat, err := file.Stat()
	if err != nil {
		log.Error().Err(err).Msg("Failed to stat PDF file")
		return "", err
	}

	pdfReader, err := pdf.NewReader(file, stat.Size())
	if err != nil {
		log.Error().Err(err).Msg("Failed to open PDF")
		return "", err
	}

	var textContent strings.Builder

	for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			log.Warn().Msgf("Skipping null page %d", pageNum)
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Warn().Msgf("Failed to extract text from page %d: %v", pageNum, err)
			continue
		}

		if text != "" {
			cleanedText := cleanPDFText(text)
			textContent.WriteString(cleanedText)
			textContent.WriteString("\n\n")
		}
	}

	out := strings.TrimSpace(textContent.String())
	if len(out) == 0 {
		return "", fmt.Errorf("no text extracted — PDF may be image-based or encrypted")
	}

	return out, nil
}

func cleanPDFText(s string) string {
	var result strings.Builder
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString(" ")
		}
	}

	return strings.TrimSpace(result.String())
}

func PDFToImage(pdfPath *os.File) (*image.RGBA, error) {
	doc, err := fitz.NewFromReader(pdfPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open PDF")
		return nil, err
	}
	defer doc.Close()

	img, err := doc.Image(0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image from PDF")
		return nil, err
	}

	return img, nil

}
