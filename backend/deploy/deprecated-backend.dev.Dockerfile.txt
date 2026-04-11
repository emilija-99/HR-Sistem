FROM golang:1.25

# Create app directory inside container
WORKDIR /app

# Install Air (correct module path!)
RUN go install github.com/air-verse/air@latest

# Air binary location
ENV PATH="/go/bin:${PATH}"

# Create folders expected by Air
RUN mkdir -p /app/tmp

# Start hot reload
CMD ["air"]
