FROM scratch

# Set a working directory
WORKDIR /app

# Copy the AMD64 binary from your local directory to the container
COPY main-amd64 /app/main

# Copy the index.html file to the working directory
COPY index.html /app/index.html

# Expose port 8080 to the outside world
EXPOSE 8080

# Set the entry point to run your application
CMD ["/app/main"]

