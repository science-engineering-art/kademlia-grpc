FROM docker.uclv.cu/golang:1.20-alpine

WORKDIR /kademlia

COPY ./ ./

CMD [ "go", "run", "./main.go"]

# RUN go install github.com/cosmtrek/air@v1.43.0

# RUN go clean -modcache

# COPY ./kademlia/go.mod ./

# RUN go mod download

# CMD ["air", "-c", ".air.toml"]
