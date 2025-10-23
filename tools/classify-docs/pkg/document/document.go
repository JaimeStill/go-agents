package document

import "fmt"

type ImageFormat string

const (
	PNG  ImageFormat = "png"
	JPEG ImageFormat = "jpg"
)

func (f ImageFormat) MimeType() (string, error) {
	switch f {
	case PNG:
		return "image/png", nil
	case JPEG:
		return "image/jpeg", nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", f)
	}
}

type ImageOptions struct {
	Format  ImageFormat
	Quality int
	DPI     int
}

func DefaultImageOptions() ImageOptions {
	return ImageOptions{
		Format:  PNG,
		Quality: 0,
		DPI:     300,
	}
}

type Document interface {
	PageCount() int
	ExtractPage(pageNum int) (Page, error)
	ExtractAllPages() ([]Page, error)
	Close() error
}

type Page interface {
	Number() int
	ToImage(opts ImageOptions) ([]byte, error)
}
