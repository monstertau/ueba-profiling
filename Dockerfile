### BUILD
# Base image `docker pull golang:1.13.14-alpine3.11`
FROM golang:1.19-alpine as build
# Folder in Container, /sample same level as /home
WORKDIR /building_stage

# Copy project code to Container
COPY . .

# Go build in Container
RUN go build -mod=vendor -o /building_stage/main ./main.go


FROM alpine

# Create workdir in target Container
WORKDIR /ueba_profiling

# Copy binary from `build` to target Container
COPY --from=build /building_stage/main /ueba_profiling/main
COPY --from=build /building_stage/config.yml /ueba_profiling/config.yml

# Run command
CMD /ueba_profiling/main