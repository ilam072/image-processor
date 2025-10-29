package processor

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/ilam072/image-processor/pkg/errutils"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
)

type Opts struct {
	Width  int
	Height int
}

type Processor struct {
	opts Opts
}

func NewProcessor(opts Opts) *Processor {
	return &Processor{opts: opts}
}

func (p *Processor) Resize(img io.ReadCloser) (io.ReadCloser, error) {
	decodedImage, _, err := image.Decode(img)
	if err != nil {
		return nil, errutils.Wrap("failed to decode image", err)
	}

	resizedImage := imaging.Resize(decodedImage, p.opts.Width, p.opts.Height, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resizedImage, imaging.JPEG); err != nil {
		return nil, errutils.Wrap("failed to encode resized image", err)
	}

	// Возвращаем новый io.ReadCloser
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func (p *Processor) Thumbnail(img io.ReadCloser) (io.ReadCloser, error) {
	decodedImage, _, err := image.Decode(img)
	if err != nil {
		return nil, errutils.Wrap("failed to decode image", err)
	}

	thumbnailImage := imaging.Thumbnail(decodedImage, p.opts.Width/10, p.opts.Height/10, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumbnailImage, imaging.JPEG); err != nil {
		return nil, errutils.Wrap("failed to encode thumbnail image", err)
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func (p *Processor) Watermark(img io.ReadCloser) (io.ReadCloser, error) {
	decodedImage, _, err := image.Decode(img)
	if err != nil {
		return nil, errutils.Wrap("failed to decode image", err)
	}

	bounds := decodedImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Копируем исходное изображение в RGBA
	rgba := imaging.Clone(decodedImage)

	// Текст watermark
	watermarkText := "© MyBrand"
	textColor := color.NRGBA{255, 255, 255, 220} // белый, почти непрозрачный
	bgColor := color.NRGBA{0, 0, 0, 120}         // полупрозрачный черный фон

	// Определяем размер текста
	textWidth := len(watermarkText) * 10 // простая оценка ширины
	textHeight := 20

	// Позиция в нижнем правом углу с отступом
	margin := 10
	x := width - textWidth - margin
	y := height - textHeight - margin

	// Рисуем фон под текст
	rect := image.Rect(x, y, x+textWidth, y+textHeight)
	draw.Draw(rgba, rect, &image.Uniform{bgColor}, image.Point{}, draw.Over)

	// Рисуем сам текст
	point := fixed.Point26_6{
		X: fixed.I(x + 2),
		Y: fixed.I(y + textHeight - 4),
	}
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(watermarkText)

	// Кодируем обратно в JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 95}); err != nil {
		return nil, errutils.Wrap("failed to encode watermarked image", err)
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
