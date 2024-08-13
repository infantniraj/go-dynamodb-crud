# Use the official Golang image as the base image
FROM golang:1.20

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code to the workspace
COPY . .

# Build the Go app
RUN go build -o /go-dynamodb-crud

# Expose port 8080
EXPOSE 8080

# Copy the .env file into the container
COPY .env .env

# Command to run the executable
CMD ["/go-dynamodb-crud"]

