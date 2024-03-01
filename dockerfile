# Use an official Go runtime as a parent image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /opt/blockchat

COPY ./blockchat.linux.x64 /opt/blockchat/blockchat
COPY ./input /opt/blockchat/input
ENV PATH="/opt/blockchat/:$PATH"
# Expose port 1500 to the outside world
EXPOSE 1500
# Command to run the executable
CMD ["blockchat","start"]