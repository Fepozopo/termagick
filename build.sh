#!/bin/bash
export CGO_CFLAGS="-I$(brew --prefix imagemagick)/include/ImageMagick-7 -DMAGICKCORE_HDRI_ENABLE=1 -DMAGICKCORE_QUANTUM_DEPTH=16 -DMAGICKCORE_CHANNEL_MASK_DEPTH=32"
export CGO_LDFLAGS="-L$(brew --prefix imagemagick)/lib -lMagickWand-7.Q16HDRI -lMagickCore-7.Q16HDRI"
go build -tags no_pkgconfig -o bin/
