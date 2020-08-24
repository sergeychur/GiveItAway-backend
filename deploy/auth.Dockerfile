FROM golang:alpine AS builder

WORKDIR /home/app/

COPY go.mod .

RUN  go mod download && go mod verify

COPY . .

RUN go build -o ./auth.out /home/app/cmd/auth/main.go

FROM bashell/alpine-bash AS release

WORKDIR /home/app/

COPY ./wait_for_it.sh /home/app
RUN chmod +x /home/app/wait_for_it.sh
COPY ./config_deploy.json /home/app/
COPY --from=builder /home/app/auth.out /home/app
RUN ls