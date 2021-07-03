FROM golang:alpine as base

FROM base as builder
WORKDIR /opt/vaccine

ADD .gitignore go.mod go.sum main.go ./

RUN go build
RUN ls
RUN echo 000

FROM base
WORKDIR /opt

COPY --from=builder /opt/vaccine/vaccine .

RUN chmod +x /opt/vaccine

ENTRYPOINT [ "/opt/vaccine" ]