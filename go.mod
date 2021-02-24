module github.com/go-vela/vela-s3-cache

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/go-vela/archiver v2.1.0+incompatible // indirect
	github.com/go-vela/types v0.7.3
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/minio/minio-go/v7 v7.0.10
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/urfave/cli/v2 v2.3.0
)

replace github.com/mholt/archiver => github.com/go-vela/archiver v1.1.3-0.20200811184543-d2452770f58c
