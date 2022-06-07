package common

const NotAvailable = "n/a"

type ImageAndTag struct {
	image string
	tag   string
}

func (iat ImageAndTag) Tag() string {
	return iat.tag
}

func (iat ImageAndTag) HasTag() bool {
	return iat.tag != "" && iat.tag != NotAvailable
}

func (iat ImageAndTag) Image() string {
	return iat.image
}

func NewImageAndTag(image string, tag string) ImageAndTag {
	return ImageAndTag{
		image: image,
		tag:   tag,
	}
}

func NewNotAvailableImageAndTag() ImageAndTag {
	return ImageAndTag{
		tag:   NotAvailable,
		image: NotAvailable,
	}
}
