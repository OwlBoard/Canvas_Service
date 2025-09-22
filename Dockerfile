# Stage 1: The "builder" stage
# We use a full Python image to install dependencies.
FROM python:3.11-slim as builder

# Set the working directory inside the container
WORKDIR /app

# Install system dependencies that might be needed by Python packages
RUN apt-get update && apt-get install -y --no-install-recommends build-essential

# Copy only the requirements file to leverage Docker's layer caching
COPY ./requirements.txt .

# Install the Python dependencies
RUN pip wheel --no-cache-dir --wheel-dir /app/wheels -r requirements.txt


# Stage 2: The "final" stage
# We start from a clean, lightweight Python image.
FROM python:3.11-slim

# Set the working directory
WORKDIR /app

# Copy the pre-built wheels from the "builder" stage
COPY --from=builder /app/wheels /wheels

# Copy the requirements file
COPY ./requirements.txt .

# Install the dependencies from the local wheels, which is much faster
RUN pip install --no-cache /wheels/*

# Copy your application source code into the container
# This assumes your code will be inside an "app/" directory
COPY ./app ./app

# Expose the port that the Uvicorn server will run on
EXPOSE 8000

# The command to run your application
# This tells uvicorn to run the "app" object inside the "main.py" file.
# --host 0.0.0.0 makes it accessible from outside the container.
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]