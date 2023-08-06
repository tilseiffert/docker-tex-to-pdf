# docker-tex-to-pdf
Simple (but not small) Docker which takes a latex file and compiles and converts it to PDF/A

## Build docker

1. First build base Docker image
    - Run `docker build -f Dockerfile.base -t tex-to-pdfa-base:12-slim .`
    - `12-slim` is equivalent to the os-image
    - This base Docker is primarily intended for better caching, as this Docker image will be quite large (ca. 5 GB and around several minutes building due to download)

2. Build the app Docker image
    - Run `docker build -f Dockerfile.app -t tex-to-pdfa .`

## Usage of docker image

1. Prepare source-dir
    - The TeX-file is to be expected as `main.tex`
    - The dir must contain all necessary files

2. Run Docker
    - Map/mount the source-dir to `/data`
    - Example: `docker run -v "$(pwd)/test":/data tex-to-pdfa`

3. Enjoy PDF/A file
    - Final file is put into source-dir named `main.pdf`