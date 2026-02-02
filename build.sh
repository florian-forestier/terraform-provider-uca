#!/bin/bash
export VERSION="1.0.2"

mkdir build

GOOS=darwin GOARCH=amd64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_darwin_amd64.zip terraform-provider-uca_v$VERSION

GOOS=darwin GOARCH=arm64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_darwin_arm64.zip terraform-provider-uca_v$VERSION

GOOS=freebsd GOARCH=386 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_freebsd_386.zip terraform-provider-uca_v$VERSION

GOOS=freebsd GOARCH=amd64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_freebsd_amd64.zip terraform-provider-uca_v$VERSION

GOOS=freebsd GOARCH=arm go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_freebsd_arm.zip terraform-provider-uca_v$VERSION

GOOS=freebsd GOARCH=arm64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_freebsd_arm64.zip terraform-provider-uca_v$VERSION

GOOS=linux GOARCH=386 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_linux_386.zip terraform-provider-uca_v$VERSION

GOOS=linux GOARCH=amd64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_linux_amd64.zip terraform-provider-uca_v$VERSION

GOOS=linux GOARCH=arm go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_linux_arm.zip terraform-provider-uca_v$VERSION

GOOS=linux GOARCH=arm64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_linux_arm64.zip terraform-provider-uca_v$VERSION

GOOS=windows GOARCH=386 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_windows_386.zip terraform-provider-uca_v$VERSION

GOOS=windows GOARCH=amd64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_windows_amd64.zip terraform-provider-uca_v$VERSION

GOOS=windows GOARCH=arm go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_windows_arm.zip terraform-provider-uca_v$VERSION

GOOS=windows GOARCH=arm64 go build -o terraform-provider-uca_v$VERSION .
zip build/terraform-provider-uca_${VERSION}_windows_arm64.zip terraform-provider-uca_v$VERSION

rm terraform-provider-uca_v$VERSION

cd build
echo "$(sha256sum *)" > terraform-provider-uca_${VERSION}_SHA256SUMS