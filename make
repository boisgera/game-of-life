#!/bin/env bash
docker run --rm --workdir=/sandbox --volume $(pwd):/sandbox boisgera/tinygo make
