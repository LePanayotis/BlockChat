# Use an official Go runtime as a parent image
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /opt/blockchat/src

# Copy the current directory contents into the container at /app
COPY ./main.go /opt/blockchat/src/main.go
COPY ./go.mod /opt/blockchat/src/go.mod
COPY ./blockchat /opt/blockchat/src/blockchat

# Download and install any required dependencies
RUN go mod tidy
RUN go mod download

# Build the Go application
RUN go build -o /opt/blockchat/blockchat .
RUN rm -rf /opt/blockchat/src
COPY ./input /opt/blockchat/input
ENV PATH="/opt/blockchat/:$PATH"
# Expose port 1500 to the outside world
EXPOSE 1500
WORKDIR /opt/blockchat
# Command to run the executable
CMD ["blockchat","start"]