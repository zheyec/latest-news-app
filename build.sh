#!/bin/bash

echo start building

env GOOS=linux GOARCH=amd64 go build -mod=vendor -o lxm-news main.go
tar czvf lxm-news.tar.gz lxm-news data static templates
scp lxm-news.tar.gz root@39.96.21.121:/home/works/chenzheye
rm lxm-news lxm-news.tar.gz