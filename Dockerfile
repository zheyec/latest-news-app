FROM golang:1.12.7 as builder
WORKDIR /home/works/program/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build  -mod=vendor -o app main.go

FROM alpine:3.10
WORKDIR /home/works/program/
COPY --from=builder /home/works/program/app .
RUN apk add ca-certificates
COPY . .
EXPOSE 8000
CMD ["/home/works/program/app","-p=8000"]