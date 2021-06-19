FROM python:3.7-stretch
COPY requirements.txt /
# RUN apt-get update && apt-get -y install gcc
# RUN apk add --no-cache python3-dev libffi-dev gcc musl-dev make
RUN pip install -r /requirements.txt
COPY . /app
WORKDIR /app
CMD ["gunicorn", "--worker-class gevent",  "--worker-connections 1000", "--workers 3", "--timeout 90", "--log-level debug", "-k uvicorn.workers.UvicornWorker", "-b 0.0.0.0", "main:app"]