# Use Python base image
FROM python:3.10-slim

# Set working directory
WORKDIR /app

# Copy the relay code
COPY main.py .

# Install required Python dependencies if any (adjust as needed)
# RUN pip install -r requirements.txt

# Expose the port used for multicast communication (9999 by default)
EXPOSE 9999

# Command to start the relay
CMD ["python", "main.py"]
