FROM tensorflow/tensorflow:2.13.0

WORKDIR /root

COPY requirements.txt .
COPY train.py .

RUN pip install -r requirements.txt

ENTRYPOINT ["exec", "train.py"]

