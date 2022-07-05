package common

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

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

// TryParseImageAndTag will try to parse imageAndTagValue into a string. It will return a ImageAndTag struct with the data it could extract from the given input data.
func TryParseImageAndTag(imageAndTagValue interface{}) ImageAndTag {

	imageAndTag, ok := imageAndTagValue.(string)
	if !ok {
		log.WithField("imageAndTag", imageAndTagValue).Error("Could not convert imageAndTag to type string")
		return NewNotAvailableImageAndTag()
	}

	if imageAndTag == NotAvailable || imageAndTag == "" {
		return NewNotAvailableImageAndTag()
	}

	split := strings.Split(imageAndTag, ":")
	if len(split) == 1 {
		return NewImageAndTag(split[0], NotAvailable)
	}

	return NewImageAndTag(split[0], split[1])
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
