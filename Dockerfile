FROM golang
WORKDIR $GOPATH/src/github.com/coryschwartz/lotus-chain-notify-example
ENV FULLNODE_API_INFO=wss://api.chain.love
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /app
EXPOSE 8080
ENTRYPOINT /app
