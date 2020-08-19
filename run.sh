#!/bin/bash
docker build -t lxm-news .
docker rm --force lxm-news
docker run --name lxm-news -e ENV="DEBUG" --restart=always -p8000:8000 lxm-news