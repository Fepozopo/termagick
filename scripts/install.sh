#!/bin/bash

# Manually set the following environment variables if pkg-config is not working
# export CGO_CFLAGS="-I$(brew --prefix imagemagick)/include/ImageMagick-7 -DMAGICKCORE_HDRI_ENABLE=1 -DMAGICKCORE_QUANTUM_DEPTH=16 -DMAGICKCORE_CHANNEL_MASK_DEPTH=32"
# export CGO_LDFLAGS="-L$(brew --prefix imagemagick)/lib -lMagickWand-7.Q16HDRI -lMagickCore-7.Q16HDRI"
# go install github.com/Fepozopo/termagick@latest

export CGO_CFLAGS="$(pkg-config --cflags MagickWand-7.Q16HDRI)"
export CGO_LDFLAGS="$(pkg-config --libs MagickWand-7.Q16HDRI)"
go install github.com/Fepozopo/termagick@latest
