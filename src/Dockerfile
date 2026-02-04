FROM golang:1.23.1-bookworm

USER axial

WORKDIR /app
COPY . .

RUN apt update && apt install -y avahi-utils

RUN go build -o axial .

CMD ["./axial"]