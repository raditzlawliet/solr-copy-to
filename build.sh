DATE_ISO=$(date +"%Y-%m-%dT%H:%M:%S")
GOARCH=amd64 GOOS=linux go build -o solr-copy-to -ldflags "-X main.BuildTime=$DATE_ISO"
GOOS=windows GOARCH=amd64 go build -o solr-copy-to.exe -ldflags "-X main.BuildTime=$DATE_ISO"