# QR Code Generator for Go

## Installation

```
got get -u github.com/grkuntzmd/qrcodegen
```

This package provides a QR Code generator that supports most of ISO/IEC
18004:2006. It is based on [Nayuki](https://github.com/nayuki/QR-Code-generator) and is released
under the Apache-2 license.

## Usage

#### Simple example

```go
import "github.com/grkuntzmd/qrcodegen"

qrCode := EncodeText("Hello, World!", Medium) // Second parameter is the error correction level (Low, Medium, Quartile, High).
svg := qrCode.ToSVGString(2, false)           // First parameter is the border width in "modules" and the second is true if you
                                              // want a "DOCTYPE" line included.
```

#### More complex example

```go
import "github.com/grkuntzmd/qrcodegen"

segs := []*QRSegment{
    MakeAlphanumeric("SUDOKU://"),
    MakeNumeric("007020004930000600600300000000000050200010008006900400003700900020050001000008000"),
}
qrCode, err := EncodeSegments(segs, Low, WithAutoMask())
if err != nil {
    // Handle this.
}
svg := qrCode.ToSVGString(4, true)
```

If you want to produce an image instead of SVG, you can access the fields in the
`QRCode` structure directly to get the size and which modules should be black:

```go
type QRCode struct {
	Version                         // The QR code version, a number in the range [1, 40].
    Size                 int        // The width and height of the square QR code symbol as measured in "modules"
    ErrorCorrectionLevel ECL        // The error correction level used in this QR code (Low, Medium, Quartile, or High).
	Mask                            // The type of mask [0, 7] used in this QR code.
    Modules              [][]Module // The modules ("pixels") that make up this QR code (black = 1, white = 0)
    // Other fields are private.
}
```

The `Modules` field, indexed by row and column, is 1 if the pixels should be
black and 0 if white.