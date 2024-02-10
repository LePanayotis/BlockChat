# Use an official Go runtime as a parent image
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Download and install any required dependencies
RUN go mod download

# Build the Go application
RUN go build -o /app/bin/bcc .

ENV PATH="/app/bin/bcc:$PATH"

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["bcc"]
