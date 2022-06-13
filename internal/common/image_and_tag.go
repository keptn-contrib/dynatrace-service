package common

const NotAvailable = "n/a"

// ImageAndTag stores the image and a tag
type ImageAndTag struct {
	image string
	tag   string
}

// Tag returns the tag
func (iat ImageAndTag) Tag() string {
	return iat.tag
}

// HasTag returns true if a tag is not empty or does not equals 'n/a'
func (iat ImageAndTag) HasTag() bool {
	return iat.tag != "" && iat.tag != NotAvailable
}

// Image returns the image
func (iat ImageAndTag) Image() string {
	return iat.image
}

// NewImageAndTag creates a new ImageAndTag instance
func NewImageAndTag(image string, tag string) ImageAndTag {
	return ImageAndTag{
		image: image,
		tag:   tag,
	}
}

// NewNotAvailableImageAndTag returns an ImageAndTag with image and tag set to 'n/a'
func NewNotAvailableImageAndTag() ImageAndTag {
	return ImageAndTag{
		tag:   NotAvailable,
		image: NotAvailable,
	}
}
