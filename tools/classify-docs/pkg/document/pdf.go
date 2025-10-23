package document

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFDocument struct {
	path      string
	ctx       *model.Context
	pageCount int
}

func OpenPDF(path string) (*PDFDocument, error) {
	ctx, err := api.ReadContextFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	pageCount := ctx.PageCount
	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	return &PDFDocument{
		path:      path,
		ctx:       ctx,
		pageCount: pageCount,
	}, nil
}

func (d *PDFDocument) PageCount() int {
	return d.pageCount
}

func (d *PDFDocument) ExtractPage(pageNum int) (Page, error) {
	if pageNum < 1 || pageNum > d.pageCount {
		return nil, fmt.Errorf("page %d out of range [1-%d]", pageNum, d.pageCount)
	}

	return &PDFPage{
		doc:    d,
		number: pageNum,
	}, nil
}

func (d *PDFDocument) ExtractAllPages() ([]Page, error) {
	pages := make([]Page, 0, d.pageCount)

	for i := 1; i <= d.pageCount; i++ {
		page, err := d.ExtractPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract page %d: %w", i, err)
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func (d *PDFDocument) Close() error {
	d.ctx = nil
	return nil
}

type PDFPage struct {
	doc    *PDFDocument
	number int
}

func (p *PDFPage) Number() int {
	return p.number
}

func (p *PDFPage) ToImage(opts ImageOptions) ([]byte, error) {
	if opts.DPI == 0 {
		opts = DefaultImageOptions()
	}

	if opts.Format == JPEG {
		if opts.Quality == 0 {
			opts.Quality = 85
		}
		if opts.Quality < 1 || opts.Quality > 100 {
			return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", opts.Quality)
		}
	}

	var ext string
	switch opts.Format {
	case PNG:
		ext = "png"
	case JPEG:
		ext = "jpg"
	default:
		return nil, fmt.Errorf("unsupported image format: %s", opts.Format)
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("page-%d-*.%s", p.number, ext))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	pageIndex := p.number - 1
	inputSpec := fmt.Sprintf("%s[%d]", p.doc.path, pageIndex)

	args := []string{
		"-density", fmt.Sprintf("%d", opts.DPI),
		inputSpec,
		"-background", "white",
		"-flatten",
	}

	if opts.Format == JPEG {
		args = append(args, "-quality", fmt.Sprintf("%d", opts.Quality))
	}

	args = append(args, tmpPath)

	cmd := exec.Command("magick", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"imagemagick failed for page %d: %w\nOutput: %s",
			p.number, err, string(output),
		)
	}

	imgData, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated image: %w", err)
	}

	return imgData, nil
}
