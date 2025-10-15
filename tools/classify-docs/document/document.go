package document

type ImageFormat string

const (
	PNG  ImageFormat = "png"
	JPEG ImageFormat = "jpg"
)

type ImageOptions struct {
	Format  ImageFormat
	Quality int
	DPI     int
}

func DefaultImageOptions() ImageOptions {
	return ImageOptions{
		Format:  PNG,
		Quality: 0,
		DPI:     150,
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
