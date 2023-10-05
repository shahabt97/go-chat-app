# Use the official Go image as the base image
FROM golang:1.21

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY . .

# Build the Go application inside the container
RUN go build -o main

# Expose the port your application listens on
EXPOSE 8080

# Command to run your application
CMD ["./main"]
