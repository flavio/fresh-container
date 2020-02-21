FROM golang:1.13-buster

COPY . /fresh-container
WORKDIR /fresh-container
RUN make build

FROM debian:buster-slim
RUN useradd -d /app web

COPY --from=0 /fresh-container/fresh-container /app/

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
WORKDIR /app
RUN chown -R web:web *

ENTRYPOINT ["/app/fresh-container"]
CMD ["server"]
EXPOSE 5000
USER web
