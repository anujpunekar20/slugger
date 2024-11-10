# Start with a lightweight Go image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application files
COPY . .

# Download the Go module dependencies
RUN go mod download

# Build the Go application
RUN go build -o slugger

# Expose the port on which the app runs
EXPOSE 8080

# Start the application
CMD ["./slugger"]