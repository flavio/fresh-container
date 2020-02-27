FROM golang:1.13-buster

COPY . /fresh-container
WORKDIR /fresh-container
RUN make build

FROM debian:buster-slim
COPY --from=0 /fresh-container/fresh-container /app/

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
WORKDIR /app
RUN chown -R www-data:www-data *

# Install certificates to reach public registries
RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

ENTRYPOINT ["/app/fresh-container"]
CMD ["server"]
EXPOSE 5000
USER www-data
